// +build !no_templategen

// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "net/http"
import "./common"

var login_Tmpl_Phrase_ID int

// nolint
func init() {
	common.Template_login_handle = Template_login
	common.Ctemplates = append(common.Ctemplates,"login")
	common.TmplPtrMap["login"] = &common.Template_login_handle
	common.TmplPtrMap["o_login"] = Template_login
	login_Tmpl_Phrase_ID = common.RegisterTmplPhraseNames([]string{
		"menu_forums_aria",
		"menu_forums_tooltip",
		"menu_topics_aria",
		"menu_topics_tooltip",
		"menu_alert_counter_aria",
		"menu_alert_list_aria",
		"menu_account_aria",
		"menu_account_tooltip",
		"menu_profile_aria",
		"menu_profile_tooltip",
		"menu_panel_aria",
		"menu_panel_tooltip",
		"menu_logout_aria",
		"menu_logout_tooltip",
		"menu_register_aria",
		"menu_register_tooltip",
		"menu_login_aria",
		"menu_login_tooltip",
		"menu_hamburger_tooltip",
		"login_head",
		"login_account_name",
		"login_account_name",
		"login_account_password",
		"login_submit_button",
		"login_no_account",
		"footer_powered_by",
		"footer_made_with_love",
		"footer_theme_selector_aria",
	})
}

// nolint
func Template_login(tmpl_login_vars common.Page, w http.ResponseWriter) error {
	var phrases = common.GetTmplPhrasesBytes(login_Tmpl_Phrase_ID)
w.Write(header_0)
w.Write([]byte(tmpl_login_vars.Title))
w.Write(header_1)
w.Write([]byte(tmpl_login_vars.Header.Site.Name))
w.Write(header_2)
w.Write([]byte(tmpl_login_vars.Header.Theme.Name))
w.Write(header_3)
if len(tmpl_login_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_login_vars.Header.Stylesheets {
w.Write(header_4)
w.Write([]byte(item))
w.Write(header_5)
}
}
w.Write(header_6)
if len(tmpl_login_vars.Header.Scripts) != 0 {
for _, item := range tmpl_login_vars.Header.Scripts {
w.Write(header_7)
w.Write([]byte(item))
w.Write(header_8)
}
}
w.Write(header_9)
w.Write([]byte(tmpl_login_vars.CurrentUser.Session))
w.Write(header_10)
w.Write([]byte(tmpl_login_vars.Header.Site.URL))
w.Write(header_11)
if tmpl_login_vars.Header.MetaDesc != "" {
w.Write(header_12)
w.Write([]byte(tmpl_login_vars.Header.MetaDesc))
w.Write(header_13)
}
w.Write(header_14)
if !tmpl_login_vars.CurrentUser.IsSuperMod {
w.Write(header_15)
}
w.Write(header_16)
w.Write(menu_0)
w.Write(menu_1)
w.Write([]byte(tmpl_login_vars.Header.Site.ShortName))
w.Write(menu_2)
w.Write(phrases[0])
w.Write(menu_3)
w.Write(phrases[1])
w.Write(menu_4)
w.Write(phrases[2])
w.Write(menu_5)
w.Write(phrases[3])
w.Write(menu_6)
w.Write(phrases[4])
w.Write(menu_7)
w.Write(phrases[5])
w.Write(menu_8)
if tmpl_login_vars.CurrentUser.Loggedin {
w.Write(menu_9)
w.Write(phrases[6])
w.Write(menu_10)
w.Write(phrases[7])
w.Write(menu_11)
w.Write([]byte(tmpl_login_vars.CurrentUser.Link))
w.Write(menu_12)
w.Write(phrases[8])
w.Write(menu_13)
w.Write(phrases[9])
w.Write(menu_14)
w.Write(phrases[10])
w.Write(menu_15)
w.Write(phrases[11])
w.Write(menu_16)
w.Write([]byte(tmpl_login_vars.CurrentUser.Session))
w.Write(menu_17)
w.Write(phrases[12])
w.Write(menu_18)
w.Write(phrases[13])
w.Write(menu_19)
} else {
w.Write(menu_20)
w.Write(phrases[14])
w.Write(menu_21)
w.Write(phrases[15])
w.Write(menu_22)
w.Write(phrases[16])
w.Write(menu_23)
w.Write(phrases[17])
w.Write(menu_24)
}
w.Write(menu_25)
w.Write(phrases[18])
w.Write(menu_26)
w.Write(header_17)
if tmpl_login_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_18)
}
w.Write(header_19)
if len(tmpl_login_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_login_vars.Header.NoticeList {
w.Write(header_20)
w.Write([]byte(item))
w.Write(header_21)
}
}
w.Write(header_22)
w.Write(login_0)
w.Write(phrases[19])
w.Write(login_1)
w.Write(phrases[20])
w.Write(login_2)
w.Write(phrases[21])
w.Write(login_3)
w.Write(phrases[22])
w.Write(login_4)
w.Write(phrases[23])
w.Write(login_5)
w.Write(phrases[24])
w.Write(login_6)
w.Write(footer_0)
w.Write([]byte(common.BuildWidget("footer",tmpl_login_vars.Header)))
w.Write(footer_1)
w.Write(phrases[25])
w.Write(footer_2)
w.Write(phrases[26])
w.Write(footer_3)
w.Write(phrases[27])
w.Write(footer_4)
if len(tmpl_login_vars.Header.Themes) != 0 {
for _, item := range tmpl_login_vars.Header.Themes {
if !item.HideFromThemes {
w.Write(footer_5)
w.Write([]byte(item.Name))
w.Write(footer_6)
if tmpl_login_vars.Header.Theme.Name == item.Name {
w.Write(footer_7)
}
w.Write(footer_8)
w.Write([]byte(item.FriendlyName))
w.Write(footer_9)
}
}
}
w.Write(footer_10)
w.Write([]byte(common.BuildWidget("rightSidebar",tmpl_login_vars.Header)))
w.Write(footer_11)
	return nil
}
