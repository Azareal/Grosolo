package panel

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
)

// TODO: Link the usernames for successful registrations to the profiles
func LogsRegs(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "registration_logs", "logs")
	if ferr != nil {
		return ferr
	}
	logCount := c.RegLogs.Count()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(logCount, page, perPage)

	logs, err := c.RegLogs.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	llist := make([]c.PageRegLogItem, len(logs))
	for index, log := range logs {
		llist[index] = c.PageRegLogItem{log, strings.Replace(strings.TrimSuffix(log.FailureReason, "|"), "|", " | ", -1)}
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelRegLogsPage{bp, llist, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "", "", "panel_reglogs", pi})
}

// TODO: Log errors when something really screwy is going on?
// TODO: Base the slugs on the localised usernames?
func handleUnknownUser(u *c.User, e error) *c.User {
	if e != nil {
		return &c.User{Name: p.GetTmplPhrase("user_unknown"), Link: c.BuildProfileURL("unknown", 0)}
	}
	return u
}
func handleUnknownTopic(t *c.Topic, e error) *c.Topic {
	if e != nil {
		return &c.Topic{Title: p.GetTmplPhrase("topic_unknown"), Link: c.BuildTopicURL("unknown", 0)}
	}
	return t
}

// TODO: Move the log building logic into /common/ and it's own abstraction
func topicElementTypeAction(action, elementType string, elementID int, actor *c.User, topic *c.Topic) (out string) {
	if action == "delete" {
		return p.GetTmplPhrasef("panel_logs_mod_action_topic_delete", elementID, actor.Link, actor.Name)
	}
	var tbit string
	aarr := strings.Split(action, "-")
	switch aarr[0] {
	case "lock", "unlock", "stick", "unstick":
		tbit = aarr[0]
	case "move":
		if len(aarr) == 2 {
			fid, _ := strconv.Atoi(aarr[1])
			forum, err := c.Forums.Get(fid)
			if err == nil {
				return p.GetTmplPhrasef("panel_logs_mod_action_topic_move_dest", topic.Link, topic.Title, forum.Link, forum.Name, actor.Link, actor.Name)
			}
		}
		tbit = "move"
	default:
		return p.GetTmplPhrasef("panel_logs_mod_action_topic_unknown", action, elementType, actor.Link, actor.Name)
	}
	if tbit != "" {
		return p.GetTmplPhrasef("panel_logs_mod_action_topic_"+tbit, topic.Link, topic.Title, actor.Link, actor.Name)
	}
	return fmt.Sprintf(out, topic.Link, topic.Title, actor.Link, actor.Name)
}

func modlogsElementType(action, elementType string, elementID int, actor *c.User) (out string) {
	switch elementType {
	case "topic":
		topic := handleUnknownTopic(c.Topics.Get(elementID))
		out = topicElementTypeAction(action, elementType, elementID, actor, topic)
	case "user":
		targetUser := handleUnknownUser(c.Users.Get(elementID))
		out = p.GetTmplPhrasef("panel_logs_mod_action_user_"+action, targetUser.Link, targetUser.Name, actor.Link, actor.Name)
	case "reply":
		if action == "delete" {
			topic := handleUnknownTopic(c.TopicByReplyID(elementID))
			out = p.GetTmplPhrasef("panel_logs_mod_action_reply_delete", topic.Link, topic.Title, actor.Link, actor.Name)
		}
	case "profile-reply":
		if action == "delete" {
			// TODO: Optimise this
			var profile *c.User
			profileReply, err := c.Prstore.Get(elementID)
			if err != nil {
				profile = &c.User{Name: p.GetTmplPhrase("user_unknown"), Link: c.BuildProfileURL("unknown", 0)}
			} else {
				profile = handleUnknownUser(c.Users.Get(profileReply.ParentID))
			}
			out = p.GetTmplPhrasef("panel_logs_mod_action_profile_reply_delete", profile.Link, profile.Name, actor.Link, actor.Name)
		}
	}
	if out == "" {
		out = p.GetTmplPhrasef("panel_logs_mod_action_unknown", action, elementType, actor.Link, actor.Name)
	}
	return out
}

func adminlogsElementType(action, elementType string, elementID int, actor *c.User, extra string) (out string) {
	switch elementType {
	// TODO: Record more detail for this, e.g. which field/s was changed
	case "user":
		tu := handleUnknownUser(c.Users.Get(elementID))
		out = p.GetTmplPhrasef("panel_logs_admin_action_user_"+action, tu.Link, tu.Name, actor.Link, actor.Name)
	case "group":
		g, err := c.Groups.Get(elementID)
		if err != nil {
			g = &c.Group{Name: p.GetTmplPhrase("group_unknown")}
		}
		out = p.GetTmplPhrasef("panel_logs_admin_action_group_"+action, "/panel/groups/edit/"+strconv.Itoa(g.ID), g.Name, actor.Link, actor.Name)
	case "group_promotion":
		out = p.GetTmplPhrasef("panel_logs_admin_action_group_promotion_"+action, actor.Link, actor.Name)
	case "forum":
		f, err := c.Forums.Get(elementID)
		if err != nil {
			f = &c.Forum{Name: p.GetTmplPhrase("forum_unknown")}
		}
		if action == "reorder" {
			out = p.GetTmplPhrasef("panel_logs_admin_action_forum_reorder", actor.Link, actor.Name)
		} else {
			out = p.GetTmplPhrasef("panel_logs_admin_action_forum_"+action, "/panel/forums/edit/"+strconv.Itoa(f.ID), f.Name, actor.Link, actor.Name)
		}
	case "page":
		pp, err := c.Pages.Get(elementID)
		if err != nil {
			pp = &c.CustomPage{Name: p.GetTmplPhrase("page_unknown")}
		}
		out = p.GetTmplPhrasef("panel_logs_admin_action_page_"+action, "/panel/pages/edit/"+strconv.Itoa(pp.ID), pp.Name, actor.Link, actor.Name)
	case "setting":
		s, err := c.SettingBox.Load().(c.SettingMap).BypassGet(action)
		if err != nil {
			s = &c.Setting{Name: p.GetTmplPhrase("setting_unknown")}
		}
		out = p.GetTmplPhrasef("panel_logs_admin_action_setting_edit", "/panel/settings/edit/"+s.Name, s.Name, actor.Link, actor.Name)
	case "word_filter":
		out = p.GetTmplPhrasef("panel_logs_admin_action_word_filter_"+action, actor.Link, actor.Name)
	case "menu":
		if action == "suborder" {
			out = p.GetTmplPhrasef("panel_logs_admin_action_menu_suborder", elementID, actor.Link, actor.Name)
		}
	case "menu_item":
		out = p.GetTmplPhrasef("panel_logs_admin_action_menu_item_"+action, "/panel/themes/menus/item/edit/"+strconv.Itoa(elementID), elementID, actor.Link, actor.Name)
	case "widget":
		out = p.GetTmplPhrasef("panel_logs_admin_action_widget_"+action, "/panel/themes/widgets/", elementID, actor.Link, actor.Name)
	case "plugin":
		out = p.GetTmplPhrasef("panel_logs_admin_action_plugin_"+action, extra, actor.Link, actor.Name)
	case "backup":
		out = p.GetTmplPhrasef("panel_logs_admin_action_backup_"+action, actor.Link, actor.Name)
	}
	if out == "" {
		out = p.GetTmplPhrasef("panel_logs_admin_action_unknown", action, elementType, actor.Link, actor.Name)
	}
	return out
}

func LogsMod(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "mod_logs", "logs")
	if ferr != nil {
		return ferr
	}
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(c.ModLogs.Count(), page, perPage)

	logs, err := c.ModLogs.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	llist := make([]c.PageLogItem, len(logs))
	for index, log := range logs {
		actor := handleUnknownUser(c.Users.Get(log.ActorID))
		action := modlogsElementType(log.Action, log.ElementType, log.ElementID, actor)
		llist[index] = c.PageLogItem{Action: template.HTML(action), IP: log.IP, DoneAt: log.DoneAt}
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelLogsPage{bp, llist, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "", "", "panel_modlogs", pi})
}

func LogsAdmin(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "admin_logs", "logs")
	if ferr != nil {
		return ferr
	}
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(c.AdminLogs.Count(), page, perPage)

	logs, err := c.AdminLogs.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	llist := make([]c.PageLogItem, len(logs))
	for index, log := range logs {
		actor := handleUnknownUser(c.Users.Get(log.ActorID))
		action := adminlogsElementType(log.Action, log.ElementType, log.ElementID, actor, log.Extra)
		llist[index] = c.PageLogItem{Action: template.HTML(action), IP: log.IP, DoneAt: log.DoneAt}
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelLogsPage{bp, llist, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "", "", "panel_adminlogs", pi})
}
