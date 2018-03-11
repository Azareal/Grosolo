// +build !no_templategen

// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "net/http"
import "./common"
import "strconv"

var topic_alt_Tmpl_Phrase_ID int

// nolint
func init() {
	common.Template_topic_alt_handle = Template_topic_alt
	common.Ctemplates = append(common.Ctemplates,"topic_alt")
	common.TmplPtrMap["topic_alt"] = &common.Template_topic_alt_handle
	common.TmplPtrMap["o_topic_alt"] = Template_topic_alt
	topic_alt_Tmpl_Phrase_ID = common.RegisterTmplPhraseNames([]string{
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
		"paginator_prev_page_aria",
		"paginator_less_than",
		"paginator_next_page_aria",
		"paginator_greater_than",
		"topic_opening_post_aria",
		"status_closed_tooltip",
		"topic_status_closed_aria",
		"topic_title_input_aria",
		"topic_update_button",
		"topic_userinfo_aria",
		"topic_level_prefix",
		"topic_poll_vote",
		"topic_poll_results",
		"topic_poll_cancel",
		"topic_opening_post_aria",
		"topic_userinfo_aria",
		"topic_level_prefix",
		"topic_like_aria",
		"topic_edit_aria",
		"topic_delete_aria",
		"topic_unlock_aria",
		"topic_lock_aria",
		"topic_unpin_aria",
		"topic_pin_aria",
		"topic_ip_full_tooltip",
		"topic_ip_full_aria",
		"topic_report_aria",
		"topic_like_count_aria",
		"topic_ip_full_tooltip",
		"topic_userinfo_aria",
		"topic_level_prefix",
		"topic_post_like_aria",
		"topic_post_edit_aria",
		"topic_post_delete_aria",
		"topic_ip_full_tooltip",
		"topic_ip_full_aria",
		"topic_report_aria",
		"topic_post_like_count_tooltip",
		"topic_your_information",
		"topic_level_prefix",
		"topic_reply_aria",
		"topic_reply_content_alt",
		"topic_reply_add_poll_option",
		"topic_reply_button",
		"topic_reply_add_poll_button",
		"topic_reply_add_file_button",
		"footer_powered_by",
		"footer_made_with_love",
		"footer_theme_selector_aria",
	})
}

// nolint
func Template_topic_alt(tmpl_topic_alt_vars common.TopicPage, w http.ResponseWriter) error {
	var phrases = common.GetTmplPhrasesBytes(topic_alt_Tmpl_Phrase_ID)
w.Write(header_0)
w.Write([]byte(tmpl_topic_alt_vars.Title))
w.Write(header_1)
w.Write([]byte(tmpl_topic_alt_vars.Header.Site.Name))
w.Write(header_2)
w.Write([]byte(tmpl_topic_alt_vars.Header.Theme.Name))
w.Write(header_3)
if len(tmpl_topic_alt_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Stylesheets {
w.Write(header_4)
w.Write([]byte(item))
w.Write(header_5)
}
}
w.Write(header_6)
if len(tmpl_topic_alt_vars.Header.Scripts) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Scripts {
w.Write(header_7)
w.Write([]byte(item))
w.Write(header_8)
}
}
w.Write(header_9)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(header_10)
w.Write([]byte(tmpl_topic_alt_vars.Header.Site.URL))
w.Write(header_11)
if tmpl_topic_alt_vars.Header.MetaDesc != "" {
w.Write(header_12)
w.Write([]byte(tmpl_topic_alt_vars.Header.MetaDesc))
w.Write(header_13)
}
w.Write(header_14)
if !tmpl_topic_alt_vars.CurrentUser.IsSuperMod {
w.Write(header_15)
}
w.Write(header_16)
w.Write(menu_0)
w.Write(menu_1)
w.Write([]byte(tmpl_topic_alt_vars.Header.Site.ShortName))
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
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
w.Write(menu_9)
w.Write(phrases[6])
w.Write(menu_10)
w.Write(phrases[7])
w.Write(menu_11)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Link))
w.Write(menu_12)
w.Write(phrases[8])
w.Write(menu_13)
w.Write(phrases[9])
w.Write(menu_14)
w.Write(phrases[10])
w.Write(menu_15)
w.Write(phrases[11])
w.Write(menu_16)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
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
if tmpl_topic_alt_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_18)
}
w.Write(header_19)
if len(tmpl_topic_alt_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.NoticeList {
w.Write(header_20)
w.Write([]byte(item))
w.Write(header_21)
}
}
w.Write(header_22)
if tmpl_topic_alt_vars.Page > 1 {
w.Write(topic_alt_0)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_1)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Page - 1)))
w.Write(topic_alt_2)
w.Write(phrases[19])
w.Write(topic_alt_3)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_4)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Page - 1)))
w.Write(topic_alt_5)
w.Write(phrases[20])
w.Write(topic_alt_6)
}
if tmpl_topic_alt_vars.LastPage != tmpl_topic_alt_vars.Page {
w.Write(topic_alt_7)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_8)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Page + 1)))
w.Write(topic_alt_9)
w.Write(phrases[21])
w.Write(topic_alt_10)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_11)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Page + 1)))
w.Write(topic_alt_12)
w.Write(phrases[22])
w.Write(topic_alt_13)
}
w.Write(topic_alt_14)
w.Write(phrases[23])
w.Write(topic_alt_15)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_16)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_17)
if tmpl_topic_alt_vars.Topic.Sticky {
w.Write(topic_alt_18)
} else {
if tmpl_topic_alt_vars.Topic.IsClosed {
w.Write(topic_alt_19)
}
}
w.Write(topic_alt_20)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Title))
w.Write(topic_alt_21)
if tmpl_topic_alt_vars.Topic.IsClosed {
w.Write(topic_alt_22)
w.Write(phrases[24])
w.Write(topic_alt_23)
w.Write(phrases[25])
w.Write(topic_alt_24)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditTopic {
w.Write(topic_alt_25)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Title))
w.Write(topic_alt_26)
w.Write(phrases[26])
w.Write(topic_alt_27)
w.Write(phrases[27])
w.Write(topic_alt_28)
}
w.Write(topic_alt_29)
if tmpl_topic_alt_vars.Poll.ID > 0 {
w.Write(topic_alt_30)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Poll.ID)))
w.Write(topic_alt_31)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Poll.ID)))
w.Write(topic_alt_32)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_33)
w.Write(phrases[28])
w.Write(topic_alt_34)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Avatar))
w.Write(topic_alt_35)
w.Write([]byte(tmpl_topic_alt_vars.Topic.UserLink))
w.Write(topic_alt_36)
w.Write([]byte(tmpl_topic_alt_vars.Topic.CreatedByName))
w.Write(topic_alt_37)
if tmpl_topic_alt_vars.Topic.Tag != "" {
w.Write(topic_alt_38)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Tag))
w.Write(topic_alt_39)
} else {
w.Write(topic_alt_40)
w.Write(phrases[29])
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.Level)))
w.Write(topic_alt_41)
}
w.Write(topic_alt_42)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Poll.ID)))
w.Write(topic_alt_43)
if len(tmpl_topic_alt_vars.Poll.QuickOptions) != 0 {
for _, item := range tmpl_topic_alt_vars.Poll.QuickOptions {
w.Write(topic_alt_44)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Poll.ID)))
w.Write(topic_alt_45)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_46)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_47)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_48)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_49)
w.Write([]byte(item.Value))
w.Write(topic_alt_50)
}
}
w.Write(topic_alt_51)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Poll.ID)))
w.Write(topic_alt_52)
w.Write(phrases[30])
w.Write(topic_alt_53)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Poll.ID)))
w.Write(topic_alt_54)
w.Write(phrases[31])
w.Write(topic_alt_55)
w.Write(phrases[32])
w.Write(topic_alt_56)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Poll.ID)))
w.Write(topic_alt_57)
}
w.Write(topic_alt_58)
w.Write(phrases[33])
w.Write(topic_alt_59)
w.Write(phrases[34])
w.Write(topic_alt_60)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Avatar))
w.Write(topic_alt_61)
w.Write([]byte(tmpl_topic_alt_vars.Topic.UserLink))
w.Write(topic_alt_62)
w.Write([]byte(tmpl_topic_alt_vars.Topic.CreatedByName))
w.Write(topic_alt_63)
if tmpl_topic_alt_vars.Topic.Tag != "" {
w.Write(topic_alt_64)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Tag))
w.Write(topic_alt_65)
} else {
w.Write(topic_alt_66)
w.Write(phrases[35])
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.Level)))
w.Write(topic_alt_67)
}
w.Write(topic_alt_68)
w.Write([]byte(tmpl_topic_alt_vars.Topic.ContentHTML))
w.Write(topic_alt_69)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Content))
w.Write(topic_alt_70)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
if tmpl_topic_alt_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_alt_71)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_72)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_73)
w.Write(phrases[36])
w.Write(topic_alt_74)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditTopic {
w.Write(topic_alt_75)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_76)
w.Write(phrases[37])
w.Write(topic_alt_77)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.DeleteTopic {
w.Write(topic_alt_78)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_79)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_80)
w.Write(phrases[38])
w.Write(topic_alt_81)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.CloseTopic {
if tmpl_topic_alt_vars.Topic.IsClosed {
w.Write(topic_alt_82)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_83)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_84)
w.Write(phrases[39])
w.Write(topic_alt_85)
} else {
w.Write(topic_alt_86)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_87)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_88)
w.Write(phrases[40])
w.Write(topic_alt_89)
}
}
if tmpl_topic_alt_vars.CurrentUser.Perms.PinTopic {
if tmpl_topic_alt_vars.Topic.Sticky {
w.Write(topic_alt_90)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_91)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_92)
w.Write(phrases[41])
w.Write(topic_alt_93)
} else {
w.Write(topic_alt_94)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_95)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_96)
w.Write(phrases[42])
w.Write(topic_alt_97)
}
}
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_98)
w.Write([]byte(tmpl_topic_alt_vars.Topic.IPAddress))
w.Write(topic_alt_99)
w.Write(phrases[43])
w.Write(topic_alt_100)
w.Write(phrases[44])
w.Write(topic_alt_101)
}
w.Write(topic_alt_102)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_103)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_104)
w.Write(phrases[45])
w.Write(topic_alt_105)
}
w.Write(topic_alt_106)
if tmpl_topic_alt_vars.Topic.LikeCount > 0 {
w.Write(topic_alt_107)
}
w.Write(topic_alt_108)
if tmpl_topic_alt_vars.Topic.LikeCount > 0 {
w.Write(topic_alt_109)
w.Write(phrases[46])
w.Write(topic_alt_110)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.LikeCount)))
w.Write(topic_alt_111)
}
w.Write(topic_alt_112)
w.Write([]byte(tmpl_topic_alt_vars.Topic.RelativeCreatedAt))
w.Write(topic_alt_113)
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_114)
w.Write([]byte(tmpl_topic_alt_vars.Topic.IPAddress))
w.Write(topic_alt_115)
w.Write(phrases[47])
w.Write(topic_alt_116)
w.Write([]byte(tmpl_topic_alt_vars.Topic.IPAddress))
w.Write(topic_alt_117)
}
w.Write(topic_alt_118)
if len(tmpl_topic_alt_vars.ItemList) != 0 {
for _, item := range tmpl_topic_alt_vars.ItemList {
w.Write(topic_alt_119)
if item.ActionType != "" {
w.Write(topic_alt_120)
}
w.Write(topic_alt_121)
w.Write(phrases[48])
w.Write(topic_alt_122)
w.Write([]byte(item.Avatar))
w.Write(topic_alt_123)
w.Write([]byte(item.UserLink))
w.Write(topic_alt_124)
w.Write([]byte(item.CreatedByName))
w.Write(topic_alt_125)
if item.Tag != "" {
w.Write(topic_alt_126)
w.Write([]byte(item.Tag))
w.Write(topic_alt_127)
} else {
w.Write(topic_alt_128)
w.Write(phrases[49])
w.Write([]byte(strconv.Itoa(item.Level)))
w.Write(topic_alt_129)
}
w.Write(topic_alt_130)
if item.ActionType != "" {
w.Write(topic_alt_131)
}
w.Write(topic_alt_132)
if item.ActionType != "" {
w.Write(topic_alt_133)
w.Write([]byte(item.ActionIcon))
w.Write(topic_alt_134)
w.Write([]byte(item.ActionType))
w.Write(topic_alt_135)
} else {
w.Write(topic_alt_136)
w.Write([]byte(item.ContentHtml))
w.Write(topic_alt_137)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
if tmpl_topic_alt_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_alt_138)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_139)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_140)
w.Write(phrases[50])
w.Write(topic_alt_141)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditReply {
w.Write(topic_alt_142)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_143)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_144)
w.Write(phrases[51])
w.Write(topic_alt_145)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.DeleteReply {
w.Write(topic_alt_146)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_147)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_148)
w.Write(phrases[52])
w.Write(topic_alt_149)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_150)
w.Write([]byte(item.IPAddress))
w.Write(topic_alt_151)
w.Write(phrases[53])
w.Write(topic_alt_152)
w.Write(phrases[54])
w.Write(topic_alt_153)
}
w.Write(topic_alt_154)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_155)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_156)
w.Write(phrases[55])
w.Write(topic_alt_157)
}
w.Write(topic_alt_158)
if item.LikeCount > 0 {
w.Write(topic_alt_159)
}
w.Write(topic_alt_160)
if item.LikeCount > 0 {
w.Write(topic_alt_161)
w.Write(phrases[56])
w.Write(topic_alt_162)
w.Write([]byte(strconv.Itoa(item.LikeCount)))
w.Write(topic_alt_163)
}
w.Write(topic_alt_164)
w.Write([]byte(item.RelativeCreatedAt))
w.Write(topic_alt_165)
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_166)
w.Write([]byte(item.IPAddress))
w.Write(topic_alt_167)
w.Write([]byte(item.IPAddress))
w.Write(topic_alt_168)
}
w.Write(topic_alt_169)
}
w.Write(topic_alt_170)
}
}
w.Write(topic_alt_171)
if tmpl_topic_alt_vars.CurrentUser.Perms.CreateReply {
w.Write(topic_alt_172)
w.Write(phrases[57])
w.Write(topic_alt_173)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Avatar))
w.Write(topic_alt_174)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Link))
w.Write(topic_alt_175)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Name))
w.Write(topic_alt_176)
if tmpl_topic_alt_vars.CurrentUser.Tag != "" {
w.Write(topic_alt_177)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Tag))
w.Write(topic_alt_178)
} else {
w.Write(topic_alt_179)
w.Write(phrases[58])
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.CurrentUser.Level)))
w.Write(topic_alt_180)
}
w.Write(topic_alt_181)
w.Write(phrases[59])
w.Write(topic_alt_182)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_183)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_184)
w.Write(phrases[60])
w.Write(topic_alt_185)
w.Write(phrases[61])
w.Write(topic_alt_186)
w.Write(phrases[62])
w.Write(topic_alt_187)
w.Write(phrases[63])
w.Write(topic_alt_188)
if tmpl_topic_alt_vars.CurrentUser.Perms.UploadFiles {
w.Write(topic_alt_189)
w.Write(phrases[64])
w.Write(topic_alt_190)
}
w.Write(topic_alt_191)
}
w.Write(topic_alt_192)
w.Write(footer_0)
w.Write([]byte(common.BuildWidget("footer",tmpl_topic_alt_vars.Header)))
w.Write(footer_1)
w.Write(phrases[65])
w.Write(footer_2)
w.Write(phrases[66])
w.Write(footer_3)
w.Write(phrases[67])
w.Write(footer_4)
if len(tmpl_topic_alt_vars.Header.Themes) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Themes {
if !item.HideFromThemes {
w.Write(footer_5)
w.Write([]byte(item.Name))
w.Write(footer_6)
if tmpl_topic_alt_vars.Header.Theme.Name == item.Name {
w.Write(footer_7)
}
w.Write(footer_8)
w.Write([]byte(item.FriendlyName))
w.Write(footer_9)
}
}
}
w.Write(footer_10)
w.Write([]byte(common.BuildWidget("rightSidebar",tmpl_topic_alt_vars.Header)))
w.Write(footer_11)
	return nil
}
