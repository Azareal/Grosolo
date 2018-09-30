package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"./common"
	"github.com/Azareal/gopsutil/mem"
)

// We're trying to reduce the amount of boilerplate in here, so I added these two functions, they might wind up circulating outside this file in the future
func panelSuccessRedirect(dest string, w http.ResponseWriter, r *http.Request, isJs bool) common.RouteError {
	if !isJs {
		http.Redirect(w, r, dest, http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}
func panelRenderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, user common.User, pi interface{}) common.RouteError {
	if common.RunPreRenderHook("pre_render_"+tmplName, w, r, &user, pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, tmplName+".html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelDashboard(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("panel_dashboard")

	// We won't calculate this on the spot anymore, as the system doesn't seem to like it if we do multiple fetches simultaneously. Should we constantly calculate this on a background thread? Perhaps, the watchdog to scale back heavy features under load? One plus side is that we'd get immediate CPU percentages here instead of waiting it to kick in with WebSockets
	var cpustr = "Unknown"
	var cpuColour string

	lessThanSwitch := func(number int, lowerBound int, midBound int) string {
		switch {
		case number < lowerBound:
			return "stat_green"
		case number < midBound:
			return "stat_orange"
		}
		return "stat_red"
	}

	var ramstr, ramColour string
	memres, err := mem.VirtualMemory()
	if err != nil {
		ramstr = "Unknown"
	} else {
		totalCount, totalUnit := common.ConvertByteUnit(float64(memres.Total))
		usedCount := common.ConvertByteInUnit(float64(memres.Total-memres.Available), totalUnit)

		// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
		var totstr string
		if (totalCount - float64(int(totalCount))) > 0.85 {
			usedCount += 1.0 - (totalCount - float64(int(totalCount)))
			totstr = strconv.Itoa(int(totalCount) + 1)
		} else {
			totstr = fmt.Sprintf("%.1f", totalCount)
		}

		if usedCount > totalCount {
			usedCount = totalCount
		}
		ramstr = fmt.Sprintf("%.1f", usedCount) + " / " + totstr + totalUnit

		ramperc := ((memres.Total - memres.Available) * 100) / memres.Total
		ramColour = lessThanSwitch(int(ramperc), 50, 75)
	}

	greaterThanSwitch := func(number int, lowerBound int, midBound int) string {
		switch {
		case number > midBound:
			return "stat_green"
		case number > lowerBound:
			return "stat_orange"
		}
		return "stat_red"
	}

	// TODO: Add a stat store for this?
	var intErr error
	var extractStat = func(stmt *sql.Stmt, args ...interface{}) (stat int) {
		err := stmt.QueryRow(args...).Scan(&stat)
		if err != nil && err != ErrNoRows {
			intErr = err
		}
		return stat
	}

	var postCount = extractStat(stmts.todaysPostCount)
	var postInterval = "day"
	var postColour = greaterThanSwitch(postCount, 5, 25)

	var topicCount = extractStat(stmts.todaysTopicCount)
	var topicInterval = "day"
	var topicColour = greaterThanSwitch(topicCount, 0, 8)

	var reportCount = extractStat(stmts.todaysTopicCountByForum, common.ReportForumID)
	var reportInterval = "week"

	var newUserCount = extractStat(stmts.todaysNewUserCount)
	var newUserInterval = "week"

	// Did any of the extractStats fail?
	if intErr != nil {
		return common.InternalError(intErr, w, r)
	}

	// TODO: Localise these
	var gridElements = []common.GridElement{
		// TODO: Implement a check for new versions of Gosora
		//common.GridElement{"dash-version", "v" + version.String(), 0, "grid_istat stat_green", "", "", "Gosora is up-to-date :)"},
		common.GridElement{"dash-version", "v" + version.String(), 0, "grid_istat", "", "", ""},

		common.GridElement{"dash-cpu", "CPU: " + cpustr, 1, "grid_istat " + cpuColour, "", "", "The global CPU usage of this server"},
		common.GridElement{"dash-ram", "RAM: " + ramstr, 2, "grid_istat " + ramColour, "", "", "The global RAM usage of this server"},
	}
	var addElement = func(element common.GridElement) {
		gridElements = append(gridElements, element)
	}

	if common.EnableWebsockets {
		uonline := common.WsHub.UserCount()
		gonline := common.WsHub.GuestCount()
		totonline := uonline + gonline
		reqCount := 0

		var onlineColour = greaterThanSwitch(totonline, 3, 10)
		var onlineGuestsColour = greaterThanSwitch(gonline, 1, 10)
		var onlineUsersColour = greaterThanSwitch(uonline, 1, 5)

		totonline, totunit := common.ConvertFriendlyUnit(totonline)
		uonline, uunit := common.ConvertFriendlyUnit(uonline)
		gonline, gunit := common.ConvertFriendlyUnit(gonline)

		addElement(common.GridElement{"dash-totonline", strconv.Itoa(totonline) + totunit + " online", 3, "grid_stat " + onlineColour, "", "", "The number of people who are currently online"})
		addElement(common.GridElement{"dash-gonline", strconv.Itoa(gonline) + gunit + " guests online", 4, "grid_stat " + onlineGuestsColour, "", "", "The number of guests who are currently online"})
		addElement(common.GridElement{"dash-uonline", strconv.Itoa(uonline) + uunit + " users online", 5, "grid_stat " + onlineUsersColour, "", "", "The number of logged-in users who are currently online"})
		addElement(common.GridElement{"dash-reqs", strconv.Itoa(reqCount) + " reqs / second", 7, "grid_stat grid_end_group " + topicColour, "", "", "The number of requests over the last 24 hours"})
	}

	addElement(common.GridElement{"dash-postsperday", strconv.Itoa(postCount) + " posts / " + postInterval, 6, "grid_stat " + postColour, "", "", "The number of new posts over the last 24 hours"})
	addElement(common.GridElement{"dash-topicsperday", strconv.Itoa(topicCount) + " topics / " + topicInterval, 7, "grid_stat " + topicColour, "", "", "The number of new topics over the last 24 hours"})
	addElement(common.GridElement{"dash-totonlineperday", "20 online / day", 8, "grid_stat stat_disabled", "", "", "Coming Soon!" /*, "The people online over the last 24 hours"*/})

	addElement(common.GridElement{"dash-searches", "8 searches / week", 9, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of searches over the last 7 days"*/})
	addElement(common.GridElement{"dash-newusers", strconv.Itoa(newUserCount) + " new users / " + newUserInterval, 10, "grid_stat", "", "", "The number of new users over the last 7 days"})
	addElement(common.GridElement{"dash-reports", strconv.Itoa(reportCount) + " reports / " + reportInterval, 11, "grid_stat", "", "", "The number of reports over the last 7 days"})

	if false {
		addElement(common.GridElement{"dash-minperuser", "2 minutes / user / week", 12, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of number of minutes spent by each active user over the last 7 days"*/})
		addElement(common.GridElement{"dash-visitorsperweek", "2 visitors / week", 13, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of unique visitors we've had over the last 7 days"*/})
		addElement(common.GridElement{"dash-postsperuser", "5 posts / user / week", 14, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of posts made by each active user over the past week"*/})
	}

	pi := common.PanelDashboardPage{&common.BasePanelPage{header, stats, "dashboard", common.ReportForumID}, gridElements}
	return panelRenderTemplate("panel_dashboard", w, r, user, &pi)
}

func routePanelThemesSetDefault(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	theme, ok := common.Themes[uname]
	if !ok {
		return common.LocalError("The theme isn't registered in the system", w, r, user)
	}
	if theme.Disabled {
		return common.LocalError("You must not enable this theme", w, r, user)
	}

	var isDefault bool
	err := stmts.isThemeDefault.QueryRow(uname).Scan(&isDefault)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	hasTheme := err != ErrNoRows
	if hasTheme {
		if isDefault {
			return common.LocalError("The theme is already active", w, r, user)
		}
		_, err = stmts.updateTheme.Exec(1, uname)
	} else {
		_, err = stmts.addTheme.Exec(uname, 1)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Make this less racey
	// TODO: Move this to common
	common.ChangeDefaultThemeMutex.Lock()
	defaultTheme := common.DefaultThemeBox.Load().(string)
	_, err = stmts.updateTheme.Exec(0, defaultTheme)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	log.Printf("Setting theme '%s' as the default theme", theme.Name)
	theme.Active = true
	common.Themes[uname] = theme

	dTheme, ok := common.Themes[defaultTheme]
	if !ok {
		return common.InternalError(errors.New("The default theme is missing"), w, r)
	}
	dTheme.Active = false
	common.Themes[defaultTheme] = dTheme

	common.DefaultThemeBox.Store(uname)
	common.ResetTemplateOverrides()
	theme.MapTemplates()
	common.ChangeDefaultThemeMutex.Unlock()

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
	return nil
}
