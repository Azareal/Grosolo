package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Polls PollStore

type PollOption struct {
	ID    int
	Value string
}

type Pollable interface {
	GetID() int
	GetTable() string
	SetPoll(pollID int) error
}

type PollStore interface {
	Get(id int) (*Poll, error)
	Exists(id int) bool
	ClearIPs() error
	Create(parent Pollable, pollType int, pollOptions map[int]string) (int, error)
	Reload(id int) error
	Count() int

	SetCache(cache PollCache)
	GetCache() PollCache
}

type DefaultPollStore struct {
	cache PollCache

	get              *sql.Stmt
	exists           *sql.Stmt
	createPoll       *sql.Stmt
	createPollOption *sql.Stmt
	delete           *sql.Stmt
	count            *sql.Stmt

	clearIPs *sql.Stmt
}

func NewDefaultPollStore(cache PollCache) (*DefaultPollStore, error) {
	acc := qgen.NewAcc()
	if cache == nil {
		cache = NewNullPollCache()
	}
	// TODO: Add an admin version of registerStmt with more flexibility?
	p := "polls"
	return &DefaultPollStore{
		cache:            cache,
		get:              acc.Select(p).Columns("parentID,parentTable,type,options,votes").Where("pollID=?").Stmt(),
		exists:           acc.Select(p).Columns("pollID").Where("pollID=?").Stmt(),
		createPoll:       acc.Insert(p).Columns("parentID,parentTable,type,options").Fields("?,?,?,?").Prepare(),
		createPollOption: acc.Insert("polls_options").Columns("pollID,option,votes").Fields("?,?,0").Prepare(),
		count:            acc.Count(p).Prepare(),

		clearIPs: acc.Update("polls_votes").Set("ip=''").Where("ip!=''").Stmt(),
	}, acc.FirstError()
}

func (s *DefaultPollStore) Exists(id int) bool {
	e := s.exists.QueryRow(id).Scan(&id)
	if e != nil && e != ErrNoRows {
		LogError(e)
	}
	return e != ErrNoRows
}

func (s *DefaultPollStore) Get(id int) (*Poll, error) {
	p, err := s.cache.Get(id)
	if err == nil {
		return p, nil
	}

	p = &Poll{ID: id}
	var optionTxt []byte
	err = s.get.QueryRow(id).Scan(&p.ParentID, &p.ParentTable, &p.Type, &optionTxt, &p.VoteCount)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(optionTxt, &p.Options)
	if err == nil {
		p.QuickOptions = s.unpackOptionsMap(p.Options)
		s.cache.Set(p)
	}
	return p, err
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// TODO: ID of 0 should always error?
func (s *DefaultPollStore) BulkGetMap(ids []int) (list map[int]*Poll, err error) {
	idCount := len(ids)
	list = make(map[int]*Poll)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := s.cache.BulkGet(ids)
	for i, sliceItem := range sliceList {
		if sliceItem != nil {
			list[sliceItem.ID] = sliceItem
		} else {
			stillHere = append(stillHere, ids[i])
		}
	}
	ids = stillHere

	// If every user is in the cache, then return immediately
	if len(ids) == 0 {
		return list, nil
	}

	idList, q := inqbuild(ids)
	rows, err := qgen.NewAcc().Select("polls").Columns("pollID,parentID,parentTable,type,options,votes").Where("pollID IN(" + q + ")").Query(idList...)
	if err != nil {
		return list, err
	}

	for rows.Next() {
		p := &Poll{ID: 0}
		var optionTxt []byte
		err := rows.Scan(&p.ID, &p.ParentID, &p.ParentTable, &p.Type, &optionTxt, &p.VoteCount)
		if err != nil {
			return list, err
		}

		err = json.Unmarshal(optionTxt, &p.Options)
		if err != nil {
			return list, err
		}
		p.QuickOptions = s.unpackOptionsMap(p.Options)
		s.cache.Set(p)

		list[p.ID] = p
	}

	// Did we miss any polls?
	if idCount > len(list) {
		var sidList string
		for _, id := range ids {
			if _, ok := list[id]; !ok {
				sidList += strconv.Itoa(id) + ","
			}
		}

		// We probably don't need this, but it might be useful in case of bugs in BulkCascadeGetMap
		if sidList == "" {
			// TODO: Bulk log this
			if Dev.DebugMode {
				log.Print("This data is sampled later in the BulkCascadeGetMap function, so it might miss the cached IDs")
				log.Print("idCount", idCount)
				log.Print("ids", ids)
				log.Print("list", list)
			}
			return list, errors.New("We weren't able to find a poll, but we don't know which one")
		}
		sidList = sidList[0 : len(sidList)-1]

		err = errors.New("Unable to find the polls with the following IDs: " + sidList)
	}

	return list, err
}

func (s *DefaultPollStore) Reload(id int) error {
	p := &Poll{ID: id}
	var optionTxt []byte
	e := s.get.QueryRow(id).Scan(&p.ParentID, &p.ParentTable, &p.Type, &optionTxt, &p.VoteCount)
	if e != nil {
		_ = s.cache.Remove(id)
		return e
	}
	e = json.Unmarshal(optionTxt, &p.Options)
	if e != nil {
		_ = s.cache.Remove(id)
		return e
	}
	p.QuickOptions = s.unpackOptionsMap(p.Options)
	_ = s.cache.Set(p)
	return nil
}

func (s *DefaultPollStore) unpackOptionsMap(rawOptions map[int]string) []PollOption {
	opts := make([]PollOption, len(rawOptions))
	for id, opt := range rawOptions {
		opts[id] = PollOption{id, opt}
	}
	return opts
}

func (s *DefaultPollStore) ClearIPs() error {
	_, e := s.clearIPs.Exec()
	return e
}

// TODO: Use a transaction for this
func (s *DefaultPollStore) Create(parent Pollable, pollType int, pollOptions map[int]string) (id int, e error) {
	// TODO: Move the option names into the polls_options table and get rid of this json sludge?
	pollOptionsTxt, e := json.Marshal(pollOptions)
	if e != nil {
		return 0, e
	}
	res, e := s.createPoll.Exec(parent.GetID(), parent.GetTable(), pollType, pollOptionsTxt)
	if e != nil {
		return 0, e
	}
	lastID, e := res.LastInsertId()
	if e != nil {
		return 0, e
	}

	for i := 0; i < len(pollOptions); i++ {
		_, e := s.createPollOption.Exec(lastID, i)
		if e != nil {
			return 0, e
		}
	}

	id = int(lastID)
	return id, parent.SetPoll(id) // TODO: Delete the poll (and options) if SetPoll fails
}

func (s *DefaultPollStore) Count() int {
	return Count(s.count)
}

func (s *DefaultPollStore) SetCache(cache PollCache) {
	s.cache = cache
}

// TODO: We're temporarily doing this so that you can do ucache != nil in getTopicUser. Refactor it.
func (s *DefaultPollStore) GetCache() PollCache {
	_, ok := s.cache.(*NullPollCache)
	if ok {
		return nil
	}
	return s.cache
}
