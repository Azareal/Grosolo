package tmpl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template/parse"
	"time"
	"unicode"
)

// TODO: Turn this file into a library
var textOverlapList = make(map[string]int)

// TODO: Stop hard-coding this here
var langPkg = "github.com/Azareal/Gosora/common/phrases"

type VarItem struct {
	Name        string
	Destination string
	Type        string
}

type VarItemReflect struct {
	Name        string
	Destination string
	Value       reflect.Value
}

type CTemplateConfig struct {
	Minify         bool
	Debug          bool
	SuperDebug     bool
	SkipHandles    bool
	SkipTmplPtrMap bool
	SkipInitBlock  bool
	PackageName    string
	DockToID       map[string]int
}

// nolint
type CTemplateSet struct {
	templateList map[string]*parse.Tree
	fileDir      string
	logDir       string
	funcMap      map[string]interface{}
	importMap    map[string]string
	//templateFragmentCount map[string]int
	fragOnce             map[string]bool
	fragmentCursor       map[string]int
	FragOut              []OutFrag
	fragBuf              []Fragment
	varList              map[string]VarItem
	localVars            map[string]map[string]VarItemReflect
	hasDispInt           bool
	localDispStructIndex int
	langIndexToName      []string
	guestOnly            bool
	memberOnly           bool
	stats                map[string]int
	//tempVars map[string]string
	config        CTemplateConfig
	baseImportMap map[string]string
	buildTags     string

	overridenTrack map[string]map[string]bool
	overridenRoots map[string]map[string]bool
	themeName      string
	perThemeTmpls  map[string]bool

	logger  *log.Logger
	loggerf *os.File
	lang    string

	fsb strings.Builder
}

func NewCTemplateSet(in string, logDir ...string) *CTemplateSet {
	var llogDir string
	if len(logDir) > 0 {
		llogDir = logDir[0]
	}
	f, err := os.OpenFile(llogDir+"tmpls-"+in+"-"+strconv.FormatInt(time.Now().Unix(), 10)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	return &CTemplateSet{
		config: CTemplateConfig{
			PackageName: "main",
		},
		logDir:         llogDir,
		baseImportMap:  map[string]string{},
		overridenRoots: map[string]map[string]bool{},
		funcMap: map[string]interface{}{
			"and":        "&&",
			"not":        "!",
			"or":         "||",
			"eq":         "==",
			"ge":         ">=",
			"gt":         ">",
			"le":         "<=",
			"lt":         "<",
			"ne":         "!=",
			"add":        "+",
			"subtract":   "-",
			"multiply":   "*",
			"divide":     "/",
			"dock":       true,
			"hasWidgets": true,
			"elapsed":    true,
			"lang":       true,
			"langf":      true,
			"level":      true,
			"bunit":      true,
			"abstime":    true,
			"reltime":    true,
			"scope":      true,
			"dyntmpl":    true,
			"ptmpl":      true,
			"js":         true,
			"index":      true,
			"flush":      true,
			"res":        true,
		},
		logger:  log.New(f, "", log.LstdFlags),
		loggerf: f,
		lang:    in,
	}
}

func (c *CTemplateSet) SetConfig(config CTemplateConfig) {
	if config.PackageName == "" {
		config.PackageName = "main"
	}
	c.config = config
}

func (c *CTemplateSet) GetConfig() CTemplateConfig {
	return c.config
}

func (c *CTemplateSet) SetBaseImportMap(importMap map[string]string) {
	c.baseImportMap = importMap
}

func (c *CTemplateSet) SetBuildTags(tags string) {
	c.buildTags = tags
}

func (c *CTemplateSet) SetOverrideTrack(overriden map[string]map[string]bool) {
	c.overridenTrack = overriden
}

func (c *CTemplateSet) GetOverridenRoots() map[string]map[string]bool {
	return c.overridenRoots
}

func (c *CTemplateSet) SetThemeName(name string) {
	c.themeName = name
}

func (c *CTemplateSet) SetPerThemeTmpls(perThemeTmpls map[string]bool) {
	c.perThemeTmpls = perThemeTmpls
}

func (c *CTemplateSet) ResetLogs(in string) {
	f, err := os.OpenFile(c.logDir+"tmpls-"+in+"-"+strconv.FormatInt(time.Now().Unix(), 10)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	c.logger = log.New(f, "", log.LstdFlags)
	c.loggerf = f
}

type SkipBlock struct {
	Frags           map[int]int
	LastCount       int
	ClosestFragSkip int
}
type Skipper struct {
	Count int
	Index int
}

type OutFrag struct {
	TmplName string
	Index    int
	Body     string
}

func (c *CTemplateSet) buildImportList() (importList string) {
	if len(c.importMap) == 0 {
		return ""
	}
	var ilsb strings.Builder
	ilsb.Grow(10 + (len(c.importMap) * 3))
	ilsb.WriteString("import (")
	for _, item := range c.importMap {
		ispl := strings.Split(item, " ")
		if len(ispl) > 1 {
			//importList += ispl[0] + " \"" + ispl[1] + "\"\n"
			ilsb.WriteString(ispl[0])
			ilsb.WriteString(" \"")
			ilsb.WriteString(ispl[1])
			ilsb.WriteString("\"\n")
		} else {
			//importList += "\"" + item + "\"\n"
			ilsb.WriteString("\"")
			ilsb.WriteString(item)
			ilsb.WriteString("\"\n")
		}
	}
	//importList += ")\n"
	ilsb.WriteString(")\n")
	return ilsb.String()
}

func (c *CTemplateSet) CompileByLoggedin(name, fileDir, expects string, expectsInt interface{}, varList map[string]VarItem, imports ...string) (stub, gout, mout string, e error) {
	c.importMap = map[string]string{}
	for index, item := range c.baseImportMap {
		c.importMap[index] = item
	}
	for _, importItem := range imports {
		c.importMap[importItem] = importItem
	}
	c.importMap["errors"] = "errors"
	importList := c.buildImportList()

	fname := strings.TrimSuffix(name, filepath.Ext(name))
	if c.themeName != "" {
		_, ok := c.perThemeTmpls[fname]
		if !ok {
			return "", "", "", nil
		}
		fname += "_" + c.themeName
	}
	c.importMap["github.com/Azareal/Gosora/common"] = "c github.com/Azareal/Gosora/common"

	c.fsb.Reset()
	stub = `package ` + c.config.PackageName + "\n" + importList + "\n"

	if !c.config.SkipInitBlock {
		//stub += "// nolint\nfunc init() {\n"
		c.fsb.WriteString("// nolint\nfunc init() {\n")
		if !c.config.SkipHandles && c.themeName == "" {
			//stub += "\tc.Tmpl_" + fname + "_handle = Tmpl_" + fname + "\n"
			c.fsb.WriteString("\tc.Tmpl_")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("_handle = Tmpl_")
			c.fsb.WriteString(fname)
			//stub += "\tc.Ctemplates = append(c.Ctemplates,\"" + fname + "\")\n"
			c.fsb.WriteString("\n\tc.Ctemplates = append(c.Ctemplates,\"")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\")\n")
		}
		if !c.config.SkipTmplPtrMap {
			//stub += "tmpl := Tmpl_" + fname + "\n"
			c.fsb.WriteString("tmpl := Tmpl_")
			c.fsb.WriteString(fname)
			//stub += "\tc.TmplPtrMap[\"" + fname + "\"] = &tmpl\n"
			c.fsb.WriteString("\n\tc.TmplPtrMap[\"")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\"] = &tmpl\n")
			//stub += "\tc.TmplPtrMap[\"o_" + fname + "\"] = tmpl\n"
			c.fsb.WriteString("\tc.TmplPtrMap[\"o_")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\"] = tmpl\n")
		}
		//stub += "}\n\n"
		c.fsb.WriteString("}\n\n")
	}
	stub += c.fsb.String()

	// TODO: Try to remove this redundant interface cast
	stub += `
// nolint
func Tmpl_` + fname + `(tmpl_i interface{}, w io.Writer) error {
	tmpl_vars, ok := tmpl_i.(` + expects + `)
	if !ok {
		return errors.New("invalid page struct value")
	}
	if tmpl_vars.CurrentUser.Loggedin {
		return Tmpl_` + fname + `_member(tmpl_i, w)
	}
	return Tmpl_` + fname + `_guest(tmpl_i, w)
}`

	c.fileDir = fileDir
	content, e := c.loadTemplate(c.fileDir, name)
	if e != nil {
		c.detail("bailing out:", e)
		return "", "", "", e
	}

	c.guestOnly = true
	gout, e = c.compile(name, content, expects, expectsInt, varList, imports...)
	if e != nil {
		return "", "", "", e
	}
	c.guestOnly = false

	c.memberOnly = true
	mout, e = c.compile(name, content, expects, expectsInt, varList, imports...)
	c.memberOnly = false

	return stub, gout, mout, e
}

func (c *CTemplateSet) Compile(name, fileDir, expects string, expectsInt interface{}, varList map[string]VarItem, imports ...string) (out string, e error) {
	if c.config.Debug {
		c.logger.Println("Compiling template '" + name + "'")
	}
	c.fileDir = fileDir
	content, e := c.loadTemplate(c.fileDir, name)
	if e != nil {
		c.detail("bailing out:", e)
		return "", e
	}

	return c.compile(name, content, expects, expectsInt, varList, imports...)
}

func (c *CTemplateSet) compile(name, content, expects string, expectsInt interface{}, varList map[string]VarItem, imports ...string) (out string, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
			if err := c.loggerf.Sync(); err != nil {
				fmt.Println(err)
			}
			log.Fatal("")
			return
		}
	}()
	//c.dumpCall("compile", name, content, expects, expectsInt, varList, imports)
	//c.detailf("c: %+v\n", c)
	c.importMap = map[string]string{}
	for index, item := range c.baseImportMap {
		c.importMap[index] = item
	}
	c.importMap["errors"] = "errors"
	for _, importItem := range imports {
		c.importMap[importItem] = importItem
	}

	c.varList = varList
	c.hasDispInt = false
	c.localDispStructIndex = 0
	c.stats = make(map[string]int)

	//tree := parse.New(name, c.funcMap)
	//treeSet := make(map[string]*parse.Tree)
	treeSet, err := parse.Parse(name, content, "{{", "}}", c.funcMap)
	if err != nil {
		return "", err
	}
	c.detail(name)
	c.detailf("treeSet: %+v\n", treeSet)

	fname := strings.TrimSuffix(name, filepath.Ext(name))
	if c.themeName != "" {
		_, ok := c.perThemeTmpls[fname]
		if !ok {
			c.detail("fname not in c.perThemeTmpls")
			c.detail("c.perThemeTmpls", c.perThemeTmpls)
			return "", nil
		}
		fname += "_" + c.themeName
	}
	if c.guestOnly {
		fname += "_guest"
	} else if c.memberOnly {
		fname += "_member"
	}

	c.detail("root overridenTrack loop")
	c.detail("fname:", fname)
	for themeName, track := range c.overridenTrack {
		c.detail("themeName:", themeName)
		c.detailf("track: %+v\n", track)
		croot, ok := c.overridenRoots[themeName]
		if !ok {
			croot = make(map[string]bool)
			c.overridenRoots[themeName] = croot
		}
		c.detailf("croot: %+v\n", croot)
		for tmplName, _ := range track {
			cname := tmplName
			if c.guestOnly {
				cname += "_guest"
			} else if c.memberOnly {
				cname += "_member"
			}
			c.detail("cname:", cname)
			if fname == cname {
				c.detail("match")
				croot[strings.TrimSuffix(strings.TrimSuffix(fname, "_guest"), "_member")] = true
			} else {
				c.detail("no match")
			}
		}
	}
	c.detailf("c.overridenRoots: %+v\n", c.overridenRoots)

	var outBuf []OutBufferFrame
	rootHold := "tmpl_" + fname + "_vars"
	//rootHold := "tmpl_vars"
	con := CContext{
		RootHolder:       rootHold,
		VarHolder:        rootHold,
		HoldReflect:      reflect.ValueOf(expectsInt),
		RootTemplateName: fname,
		TemplateName:     fname,
		OutBuf:           &outBuf,
	}

	c.templateList = map[string]*parse.Tree{}
	for nname, tree := range treeSet {
		if name == nname {
			c.templateList[fname] = tree
		} else {
			if strings.HasPrefix(nname, ".html") {
				nname = strings.TrimSuffix(nname, ".html")
			}
			c.templateList[nname] = tree
		}
	}
	c.detailf("c.templateList: %+v\n", c.templateList)

	c.localVars = make(map[string]map[string]VarItemReflect)
	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".", con.VarHolder, con.HoldReflect}
	if c.fragOnce == nil {
		c.fragOnce = make(map[string]bool)
	}
	c.fragmentCursor = map[string]int{fname: 0}
	c.fragBuf = nil
	c.langIndexToName = nil

	// TODO: Is this the first template loaded in? We really should have some sort of constructor for CTemplateSet
	//if c.templateFragmentCount == nil {
	//	c.templateFragmentCount = make(map[string]int)
	//}
	//c.detailf("c: %+v\n", c)

	c.detailf("name: %+v\n", name)
	c.detailf("fname: %+v\n", fname)
	startIndex := con.StartTemplate("")
	ttree := c.templateList[fname]
	if ttree == nil {
		panic("ttree is nil")
	}
	c.rootIterate(ttree, con)
	con.EndTemplate("")
	c.afterTemplate(con, startIndex)
	//c.templateFragmentCount[fname] = c.fragmentCursor[fname] + 1

	_, ok := c.fragOnce[fname]
	if !ok {
		c.fragOnce[fname] = true
	}
	if len(c.langIndexToName) > 0 {
		c.importMap[langPkg] = langPkg
	}
	// TODO: Simplify this logic by doing some reordering?
	if c.lang == "normal" {
		c.importMap["net/http"] = "net/http"
	}
	importList := c.buildImportList()

	c.fsb.Reset()
	//var fout string
	if c.buildTags != "" {
		//fout += "// +build " + c.buildTags + "\n\n"
		c.fsb.WriteString("// +build ")
		c.fsb.WriteString(c.buildTags)
		c.fsb.WriteString("\n\n")
	}
	//fout += "// Code generated by Gosora. More below:\n/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */\n"
	c.fsb.WriteString("// Code generated by Gosora. More below:\n/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */\n")
	//fout += "package " + c.config.PackageName + "\n" + importList + "\n"
	c.fsb.WriteString("package ")
	c.fsb.WriteString(c.config.PackageName)
	c.fsb.WriteString("\n")
	c.fsb.WriteString(importList)
	c.fsb.WriteString("\n")

	if c.lang == "js" {
		//var l string
		if len(c.langIndexToName) > 0 {
			/*var lsb strings.Builder
			lsb.Grow(len(c.langIndexToName) * (1 + 2))
			for i, name := range c.langIndexToName {
				//l += `"` + name + `"` + ",\n"
				if i == 0 {
					//l += `"` + name + `"`
					lsb.WriteRune('"')
				} else {
					//l += `,"` + name + `"`
					lsb.WriteString(`,"`)
				}
				lsb.WriteString(name)
				lsb.WriteRune('"')
			}*/
			//fout += "if(tmplInits===undefined) var tmplInits={}\n"
			c.fsb.WriteString("if(tmplInits===undefined) var tmplInits={}\n")
			//fout += "tmplInits[\"tmpl_" + fname + "\"]=[" + lsb.String() + "]"
			c.fsb.WriteString("tmplInits[\"tmpl_")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\"]=[")

			c.fsb.Grow(len(c.langIndexToName) * (1 + 2))
			for i, name := range c.langIndexToName {
				//l += `"` + name + `"` + ",\n"
				if i == 0 {
					//l += `"` + name + `"`
					c.fsb.WriteRune('"')
				} else {
					//l += `,"` + name + `"`
					c.fsb.WriteString(`,"`)
				}
				c.fsb.WriteString(name)
				c.fsb.WriteRune('"')
			}

			c.fsb.WriteString("]")
		} else {
			//fout += "if(tmplInits===undefined) var tmplInits={}\n"
			c.fsb.WriteString("if(tmplInits===undefined) var tmplInits={}\n")
			//fout += "tmplInits[\"tmpl_" + fname + "\"]=[]"
			c.fsb.WriteString("tmplInits[\"tmpl_")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\"]=[]")
		}
		/*if len(l) > 0 {
			l = "\n" + l
		}*/
	} else if !c.config.SkipInitBlock {
		if len(c.langIndexToName) > 0 {
			//fout += "var " + fname + "_tmpl_phrase_id int\n\n"
			c.fsb.WriteString("var ")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("_tmpl_phrase_id int\n\n")
			c.fsb.WriteString("var ")
			c.fsb.WriteString(fname)
			if len(c.langIndexToName) > 1 {
				//fout += "var " + fname + "_phrase_arr [" + strconv.Itoa(len(c.langIndexToName)) + "][]byte\n\n"
				c.fsb.WriteString("_phrase_arr [")
				c.fsb.WriteString(strconv.Itoa(len(c.langIndexToName)))
				c.fsb.WriteString("][]byte\n\n")
			} else {
				//fout += "var " + fname + "_phrase []byte\n\n"
				c.fsb.WriteString("_phrase []byte\n\n")
			}
		}
		//fout += "// nolint\nfunc init() {\n"
		c.fsb.WriteString("// nolint\nfunc init() {\n")

		if !c.config.SkipHandles && c.themeName == "" {
			//fout += "\tc.Tmpl_" + fname + "_handle = Tmpl_" + fname
			c.fsb.WriteString("\tc.Tmpl_")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("_handle = Tmpl_")
			c.fsb.WriteString(fname)
			//fout += "\n\tc.Ctemplates = append(c.Ctemplates,\"" + fname + "\")\n"
			c.fsb.WriteString("\n\tc.Ctemplates = append(c.Ctemplates,\"")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\")\n")
		}

		if !c.config.SkipTmplPtrMap {
			//fout += "tmpl := Tmpl_" + fname + "\n"
			c.fsb.WriteString("tmpl := Tmpl_")
			c.fsb.WriteString(fname)
			//fout += "\tc.TmplPtrMap[\"" + fname + "\"] = &tmpl\n"
			c.fsb.WriteString("\n\tc.TmplPtrMap[\"")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\"] = &tmpl\n")
			//fout += "\tc.TmplPtrMap[\"o_" + fname + "\"] = tmpl\n"
			c.fsb.WriteString("\tc.TmplPtrMap[\"o_")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("\"] = tmpl\n")
		}
		if len(c.langIndexToName) > 0 {
			//fout += "\t" + fname + "_tmpl_phrase_id = phrases.RegisterTmplPhraseNames([]string{\n"
			c.fsb.WriteString("\t")
			c.fsb.WriteString(fname)
			c.fsb.WriteString("_tmpl_phrase_id = phrases.RegisterTmplPhraseNames([]string{\n")
			for _, name := range c.langIndexToName {
				//fout += "\t\t" + `"` + name + `"` + ",\n"
				c.fsb.WriteString("\t\t\"")
				c.fsb.WriteString(name)
				c.fsb.WriteString("\",\n")
			}
			//fout += "\t})\n"
			c.fsb.WriteString("\t})\n")

			if len(c.langIndexToName) > 1 {
				/*fout += `	phrases.AddTmplIndexCallback(func(phraseSet [][]byte) {
						copy(` + fname + `_phrase_arr[:], phraseSet)
					})
				`*/
				c.fsb.WriteString(`	phrases.AddTmplIndexCallback(func(phraseSet [][]byte) {
		copy(`)
				c.fsb.WriteString(fname)
				c.fsb.WriteString(`_phrase_arr[:], phraseSet)
	})
`)
			} else {
				/*fout += `	phrases.AddTmplIndexCallback(func(phraseSet [][]byte) {
						` + fname + `_phrase = phraseSet[0]
				})
				`*/
				c.fsb.WriteString(`	phrases.AddTmplIndexCallback(func(phraseSet [][]byte) {
`)
				c.fsb.WriteString(fname)
				c.fsb.WriteString(`_phrase = phraseSet[0]
	})
`)
			}
		}
		//fout += "}\n\n"
		c.fsb.WriteString("}\n\n")
	}

	c.fsb.WriteString("// nolint\nfunc Tmpl_")
	c.fsb.WriteString(fname)
	if c.lang == "normal" {
		/*fout += "// nolint\nfunc Tmpl_" + fname + "(tmpl_i interface{}, w io.Writer) error {\n"
				fout += `tmpl_` + fname + `_vars, ok := tmpl_i.(` + expects + `)
		if !ok {
			return errors.New("invalid page struct value")
		}
		`*/
		c.fsb.WriteString("(tmpl_i interface{}, w io.Writer) error {\n")

		c.fsb.WriteString(`tmpl_`)
		c.fsb.WriteString(fname)
		c.fsb.WriteString(`_vars, ok := tmpl_i.(`)
		c.fsb.WriteString(expects)
		c.fsb.WriteString(`)
	if !ok {
		return errors.New("invalid page struct value")
	}
	var iw http.ResponseWriter
	if gzw, ok := w.(c.GzipResponseWriter); ok {
		iw = gzw.ResponseWriter
		w = gzw.Writer
	}
	_ = iw
	var tmp []byte
	_ = tmp
`)
	} else {
		//fout += "// nolint\nfunc Tmpl_" + fname + "(tmpl_" + fname + "_vars interface{}, w io.Writer) error {\n"
		c.fsb.WriteString("(tmpl_")
		c.fsb.WriteString(fname)
		c.fsb.WriteString("_vars interface{}, w io.Writer) error {\n")
		//fout += "// nolint\nfunc Tmpl_" + fname + "(tmpl_vars interface{}, w io.Writer) error {\n"
	}

	//var fsb strings.Builder
	if len(c.langIndexToName) > 0 {
		//fout += "//var plist = phrases.GetTmplPhrasesBytes(" + fname + "_tmpl_phrase_id)\n"
		c.fsb.WriteString("//var plist = phrases.GetTmplPhrasesBytes(")
		c.fsb.WriteString(fname)
		c.fsb.WriteString("_tmpl_phrase_id)\n")

		//fout += "if len(plist) > 0 {\n_ = plist[len(plist)-1]\n}\n"
		//fout += "var plist = " + fname + "_phrase_arr\n"
	}

	//var varString string
	//var vssb strings.Builder
	c.fsb.Grow(10 + 3)
	for _, varItem := range c.varList {
		//varString += "var " + varItem.Name + " " + varItem.Type + " = " + varItem.Destination + "\n"
		c.fsb.WriteString("var ")
		c.fsb.WriteString(varItem.Name)
		c.fsb.WriteRune(' ')
		c.fsb.WriteString(varItem.Type)
		c.fsb.WriteString(" = ")
		c.fsb.WriteString(varItem.Destination)
		c.fsb.WriteString("\n")
	}

	//c.fsb.WriteString(varString)
	//fout += varString
	skipped := make(map[string]*SkipBlock) // map[templateName]*SkipBlock{map[atIndexAndAfter]skipThisMuch,lastCount}

	writeTextFrame := func(tmplName string, index int) {
		out := "w.Write(" + tmplName + "_frags[" + strconv.Itoa(index) + "]" + ")\n"
		c.detail("writing ", out)
		//fout += out
		c.fsb.WriteString(out)
	}

	for fid := 0; len(outBuf) > fid; fid++ {
		fr := outBuf[fid]
		c.detail(fr.Type + " frame")
		switch {
		case fr.Type == "text":
			c.detail(fr)
			oid := fid
			c.detail("oid:", oid)
			skipBlock, ok := skipped[fr.TemplateName]
			if !ok {
				skipBlock = &SkipBlock{make(map[int]int), 0, 0}
				skipped[fr.TemplateName] = skipBlock
			}
			skip := skipBlock.LastCount
			c.detailf("skipblock %+v\n", skipBlock)
			//var count int
			for len(outBuf) > fid+1 && outBuf[fid+1].Type == "text" && outBuf[fid+1].TemplateName == fr.TemplateName {
				c.detail("pre fid:", fid)
				//count++
				next := outBuf[fid+1]
				c.detail("next frame:", next)
				c.detail("frame frag:", c.fragBuf[fr.Extra2.(int)])
				c.detail("next frag:", c.fragBuf[next.Extra2.(int)])
				c.fragBuf[fr.Extra2.(int)].Body += c.fragBuf[next.Extra2.(int)].Body
				c.fragBuf[next.Extra2.(int)].Seen = true
				fid++
				skipBlock.LastCount++
				skipBlock.Frags[fr.Extra.(int)] = skipBlock.LastCount
				c.detail("post fid:", fid)
			}
			writeTextFrame(fr.TemplateName, fr.Extra.(int)-skip)
		case fr.Type == "varsub" || fr.Type == "cvarsub":
			//fout += "w.Write(" + fr.Body + ")\n"
			c.fsb.WriteString("w.Write(")
			c.fsb.WriteString(fr.Body)
			c.fsb.WriteString(")\n")
		case fr.Type == "lang":
			//fout += "w.Write(plist[" + strconv.Itoa(fr.Extra.(int)) + "])\n"
			c.fsb.WriteString("w.Write(")
			c.fsb.WriteString(fname)
			if len(c.langIndexToName) == 1 {
				//fout += "w.Write(" + fname + "_phrase)\n"
				c.fsb.WriteString("_phrase)\n")
			} else {
				//fout += "w.Write(" + fname + "_phrase_arr[" + strconv.Itoa(fr.Extra.(int)) + "])\n"
				c.fsb.WriteString("_phrase_arr[")
				c.fsb.WriteString(strconv.Itoa(fr.Extra.(int)))
				c.fsb.WriteString("])\n")
			}
		//case fr.Type == "identifier":
		default:
			//fout += fr.Body
			c.fsb.WriteString(fr.Body)
		}
	}
	//fout += "return nil\n}\n"
	c.fsb.WriteString("return nil\n}\n")
	//fout += c.fsb.String()

	writeFrag := func(tmplName string, index int, body string) {
		//c.detail("writing ", fragmentPrefix)
		c.FragOut = append(c.FragOut, OutFrag{tmplName, index, body})
	}

	for _, frag := range c.fragBuf {
		c.detail("frag:", frag)
		if frag.Seen {
			c.detail("invisible")
			continue
		}
		// TODO: What if the same template is invoked in multiple spots in a template?
		skipBlock := skipped[frag.TemplateName]
		skip := skipBlock.Frags[skipBlock.ClosestFragSkip]
		_, ok := skipBlock.Frags[frag.Index]
		if ok {
			skipBlock.ClosestFragSkip = frag.Index
		}
		c.detailf("skipblock %+v\n", skipBlock)
		c.detail("skipping ", skip)
		index := frag.Index - skip
		if index < 0 {
			index = 0
		}
		writeFrag(frag.TemplateName, index, frag.Body)
	}

	fout := strings.Replace(c.fsb.String(), `))
w.Write([]byte(`, " + ", -1)
	fout = strings.Replace(fout, "` + `", "", -1)

	if c.config.Debug {
		for index, count := range c.stats {
			c.logger.Println(index+": ", strconv.Itoa(count))
		}
		c.logger.Println(" ")
	}
	c.detail("Output!")
	c.detail(fout)
	return fout, nil
}

func (c *CTemplateSet) rootIterate(tree *parse.Tree, con CContext) {
	c.dumpCall("rootIterate", tree, con)
	if tree.Root == nil {
		c.detailf("tree: %+v\n", tree)
		panic("tree root node is empty")
	}
	c.detail(tree.Root)
	for _, node := range tree.Root.Nodes {
		c.detail("Node:", node.String())
		c.compileSwitch(con, node)
	}
	c.retCall("rootIterate")
}

func inSlice(haystack []string, expr string) bool {
	for _, needle := range haystack {
		if needle == expr {
			return true
		}
	}
	return false
}

func (c *CTemplateSet) compileSwitch(con CContext, node parse.Node) {
	c.dumpCall("compileSwitch", con, node)
	defer c.retCall("compileSwitch")
	switch node := node.(type) {
	case *parse.ActionNode:
		c.detail("Action Node")
		if node.Pipe == nil {
			break
		}
		for _, cmd := range node.Pipe.Cmds {
			c.compileSubSwitch(con, cmd)
		}
	case *parse.IfNode:
		c.detail("If Node:")
		c.detail("node.Pipe", node.Pipe)
		var expr string
		for _, cmd := range node.Pipe.Cmds {
			c.detail("If Node Bit:", cmd)
			c.detail("Bit Type:", reflect.ValueOf(cmd).Type().Name())
			exprStep := c.compileExprSwitch(con, cmd)
			expr += exprStep
			c.detail("Expression Step:", exprStep)
		}

		c.detail("Expression:", expr)
		// Simple member / guest optimisation for now
		// TODO: Expand upon this
		userExprs, negUserExprs := buildUserExprs(con.RootHolder)
		if c.guestOnly {
			c.detail("optimising away member branch")
			if inSlice(userExprs, expr) {
				c.detail("positive conditional:", expr)
				if node.ElseList != nil {
					c.compileSwitch(con, node.ElseList)
				}
				return
			} else if inSlice(negUserExprs, expr) {
				c.detail("negative conditional:", expr)
				c.compileSwitch(con, node.List)
				return
			}
		} else if c.memberOnly {
			c.detail("optimising away guest branch")
			if (con.RootHolder + ".CurrentUser.Loggedin") == expr {
				c.detail("positive conditional:", expr)
				c.compileSwitch(con, node.List)
				return
			} else if ("!" + con.RootHolder + ".CurrentUser.Loggedin") == expr {
				c.detail("negative conditional:", expr)
				if node.ElseList != nil {
					c.compileSwitch(con, node.ElseList)
				}
				return
			}
		}

		// simple constant folding
		if expr == "true" {
			c.compileSwitch(con, node.List)
			return
		} else if expr == "false" {
			c.compileSwitch(con, node.ElseList)
			return
		}

		var startIf int
		var nilIf = strings.HasPrefix(expr, con.RootHolder) && strings.HasSuffix(expr, "!=nil")
		if nilIf {
			startIf = con.StartIfPtr("if " + expr + " {\n")
		} else {
			startIf = con.StartIf("if " + expr + " {\n")
		}
		c.compileSwitch(con, node.List)
		if node.ElseList == nil {
			c.detail("Selected Branch 1")
			con.EndIf(startIf, "}\n")
			if nilIf {
				c.afterTemplate(con, startIf)
			}
		} else {
			c.detail("Selected Branch 2")
			con.EndIf(startIf, "}")
			if nilIf {
				c.afterTemplate(con, startIf)
			}
			con.Push("startelse", " else {\n")
			c.compileSwitch(con, node.ElseList)
			con.Push("endelse", "}\n")
		}
	case *parse.ListNode:
		c.detailf("List Node: %+v\n", node)
		for _, subnode := range node.Nodes {
			c.compileSwitch(con, subnode)
		}
	case *parse.RangeNode:
		c.compileRangeNode(con, node)
	case *parse.TemplateNode:
		c.compileSubTemplate(con, node)
	case *parse.TextNode:
		c.addText(con, node.Text)
	default:
		c.unknownNode(node)
	}
}

func (c *CTemplateSet) addText(con CContext, text []byte) {
	c.dumpCall("addText", con, text)
	tmpText := bytes.TrimSpace(text)
	if len(tmpText) == 0 {
		return
	}
	nodeText := string(text)
	c.detail("con.TemplateName:", con.TemplateName)
	fragIndex := c.fragmentCursor[con.TemplateName]
	_, ok := c.fragOnce[con.TemplateName]
	c.fragBuf = append(c.fragBuf, Fragment{nodeText, con.TemplateName, fragIndex, ok})
	con.PushText(strconv.Itoa(fragIndex), fragIndex, len(c.fragBuf)-1)
	c.fragmentCursor[con.TemplateName] = fragIndex + 1
}

func (c *CTemplateSet) compileRangeNode(con CContext, node *parse.RangeNode) {
	c.dumpCall("compileRangeNode", con, node)
	defer c.retCall("compileRangeNode")
	c.detail("node.Pipe:", node.Pipe)
	var expr string
	var outVal reflect.Value
	for _, cmd := range node.Pipe.Cmds {
		c.detail("Range Bit:", cmd)
		// ! This bit is slightly suspect, hm.
		expr, outVal = c.compileReflectSwitch(con, cmd)
	}
	c.detail("Expr:", expr)
	c.detail("Range Kind Switch!")

	startIf := func(item reflect.Value, useCopy bool) {
		sIndex := con.StartIf("if len(" + expr + ")!=0 {\n")
		startIndex := con.StartLoop("for _, item := range " + expr + " {\n")
		ccon := con
		var depth string
		if ccon.VarHolder == "item" {
			depth = strings.TrimPrefix(ccon.VarHolder, "item")
			if depth != "" {
				idepth, err := strconv.Atoi(depth)
				if err != nil {
					panic(err)
				}
				depth = strconv.Itoa(idepth + 1)
			}
		}
		ccon.VarHolder = "item" + depth
		ccon.HoldReflect = item
		c.compileSwitch(ccon, node.List)
		if con.LastBufIndex() == startIndex {
			con.DiscardAndAfter(startIndex - 1)
			return
		}
		con.EndLoop("}\n")
		c.afterTemplate(con, startIndex)
		if node.ElseList != nil {
			con.EndIf(sIndex, "}")
			con.Push("startelse", " else {\n")
			if !useCopy {
				ccon = con
			}
			c.compileSwitch(ccon, node.ElseList)
			con.Push("endelse", "}\n")
		} else {
			con.EndIf(sIndex, "}\n")
		}
	}

	switch outVal.Kind() {
	case reflect.Map:
		var item reflect.Value
		for _, key := range outVal.MapKeys() {
			item = outVal.MapIndex(key)
		}
		c.detail("Range item:", item)
		if !item.IsValid() {
			c.critical("expr:", expr)
			c.critical("con.VarHolder", con.VarHolder)
			panic("item" + "^\n" + "Invalid map. Maybe, it doesn't have any entries for the template engine to analyse?")
		}
		startIf(item, true)
	case reflect.Slice:
		if outVal.Len() == 0 {
			c.critical("expr:", expr)
			c.critical("con.VarHolder", con.VarHolder)
			panic("The sample data needs at-least one or more elements for the slices. We're looking into removing this requirement at some point!")
		}
		startIf(outVal.Index(0), false)
	case reflect.Invalid:
		return
	}
}

// ! Temporary, we probably want something that is good with non-struct pointers too
// For compileSubSwitch and compileSubTemplate
func (c *CTemplateSet) skipStructPointers(cur reflect.Value, id string) reflect.Value {
	if cur.Kind() == reflect.Ptr {
		c.detail("Looping over pointer")
		for cur.Kind() == reflect.Ptr {
			cur = cur.Elem()
		}
		c.detail("Data Kind:", cur.Kind().String())
		c.detail("Field Bit:", id)
	}
	return cur
}

// For compileSubSwitch and compileSubTemplate
func (c *CTemplateSet) checkIfValid(cur reflect.Value, varHolder string, holdReflect reflect.Value, varBit string, multiline bool) {
	if !cur.IsValid() {
		c.critical("Debug Data:")
		c.critical("Holdreflect:", holdReflect)
		c.critical("Holdreflect.Kind():", holdReflect.Kind())
		if !c.config.SuperDebug {
			c.critical("cur.Kind():", cur.Kind().String())
		}
		c.critical("")
		if !multiline {
			panic(varHolder + varBit + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		panic(varBit + "^\n" + "Invalid value. Maybe, it doesn't exist?")
	}
}

func (c *CTemplateSet) compileSubSwitch(con CContext, node *parse.CommandNode) {
	c.dumpCall("compileSubSwitch", con, node)
	switch n := node.Args[0].(type) {
	case *parse.FieldNode:
		c.detail("Field Node:", n.Ident)
		/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Variable declarations are coming soon! */
		cur := con.HoldReflect

		var varBit string
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
			varBit += ".(" + cur.Type().Name() + ")"
		}

		var assLines string
		multiline := false
		for _, id := range n.Ident {
			c.detail("Data Kind:", cur.Kind().String())
			c.detail("Field Bit:", id)
			cur = c.skipStructPointers(cur, id)
			c.checkIfValid(cur, con.VarHolder, con.HoldReflect, varBit, multiline)

			c.detail("in-loop varBit:" + varBit)
			if cur.Kind() == reflect.Map {
				cur = cur.MapIndex(reflect.ValueOf(id))
				varBit += "[\"" + id + "\"]"
				cur = c.skipStructPointers(cur, id)

				if cur.Kind() == reflect.Struct || cur.Kind() == reflect.Interface {
					// TODO: Move the newVarByte declaration to the top level or to the if level, if a dispInt is only used in a particular if statement
					var dispStr, newVarByte string
					if cur.Kind() == reflect.Interface {
						dispStr = "Int"
						if !c.hasDispInt {
							newVarByte = ":"
							c.hasDispInt = true
						}
					}
					// TODO: De-dupe identical struct types rather than allocating a variable for each one
					if cur.Kind() == reflect.Struct {
						dispStr = "Struct" + strconv.Itoa(c.localDispStructIndex)
						newVarByte = ":"
						c.localDispStructIndex++
					}
					con.VarHolder = "disp" + dispStr
					varBit = con.VarHolder + " " + newVarByte + "= " + con.VarHolder + varBit + "\n"
					multiline = true
				} else {
					continue
				}
			}
			if cur.Kind() != reflect.Interface {
				cur = cur.FieldByName(id)
				varBit += "." + id
			}

			// TODO: Handle deeply nested pointers mixed with interfaces mixed with pointers better
			if cur.Kind() == reflect.Interface {
				cur = cur.Elem()
				varBit += ".("
				// TODO: Surely, there's a better way of doing this?
				if cur.Type().PkgPath() != "main" && cur.Type().PkgPath() != "" {
					c.importMap["html/template"] = "html/template"
					varBit += strings.TrimPrefix(cur.Type().PkgPath(), "html/") + "."
				}
				varBit += cur.Type().Name() + ")"
			}
			c.detail("End Cycle:", varBit)
		}

		if multiline {
			assSplit := strings.Split(varBit, "\n")
			varBit = assSplit[len(assSplit)-1]
			assSplit = assSplit[:len(assSplit)-1]
			assLines = strings.Join(assSplit, "\n") + "\n"
		}
		c.compileVarSub(con, con.VarHolder+varBit, cur, assLines, func(in string) string {
			for _, varItem := range c.varList {
				if strings.HasPrefix(in, varItem.Destination) {
					in = strings.Replace(in, varItem.Destination, varItem.Name, 1)
				}
			}
			return in
		})
	case *parse.DotNode:
		c.detail("Dot Node:", node.String())
		c.compileVarSub(con, con.VarHolder, con.HoldReflect, "", nil)
	case *parse.NilNode:
		panic("Nil is not a command x.x")
	case *parse.VariableNode:
		c.detail("Variable Node:", n.String())
		c.detail(n.Ident)
		varname, reflectVal := c.compileIfVarSub(con, n.String())
		c.compileVarSub(con, varname, reflectVal, "", nil)
	case *parse.StringNode:
		con.Push("stringnode", n.Quoted)
	case *parse.IdentifierNode:
		c.detail("Identifier Node:", node)
		c.detail("Identifier Node Args:", node.Args)
		out, outval, lit, noident := c.compileIdentSwitch(con, node)
		if noident {
			return
		} else if lit {
			con.Push("identifier", out)
			return
		}
		c.compileVarSub(con, out, outval, "", nil)
	default:
		c.unknownNode(node)
	}
}

func (c *CTemplateSet) compileExprSwitch(con CContext, node *parse.CommandNode) (out string) {
	c.dumpCall("compileExprSwitch", con, node)
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
	case *parse.FieldNode:
		if c.config.SuperDebug {
			c.logger.Println("Field Node:", n.Ident)
			for _, id := range n.Ident {
				c.logger.Println("Field Bit:", id)
			}
		}
		/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Coming Soon. */
		out = c.compileBoolSub(con, n.String())
	case *parse.ChainNode:
		c.detail("Chain Node:", n.Node)
		c.detail("Node Args:", node.Args)
	case *parse.IdentifierNode:
		c.detail("Identifier Node:", node)
		c.detail("Node Args:", node.Args)
		out = c.compileIdentSwitchN(con, node)
	case *parse.DotNode:
		out = con.VarHolder
	case *parse.VariableNode:
		c.detail("Variable Node:", n.String())
		c.detail("Node Identifier:", n.Ident)
		out, _ = c.compileIfVarSub(con, n.String())
	case *parse.NilNode:
		panic("Nil is not a command x.x")
	case *parse.PipeNode:
		c.detail("Pipe Node:", n)
		c.detail("Node Args:", node.Args)
		out += c.compileIdentSwitchN(con, node)
	default:
		c.unknownNode(firstWord)
	}
	c.retCall("compileExprSwitch", out)
	return out
}

func (c *CTemplateSet) unknownNode(n parse.Node) {
	el := reflect.ValueOf(n).Elem()
	c.logger.Println("Unknown Kind:", el.Kind())
	c.logger.Println("Unknown Type:", el.Type().Name())
	panic("I don't know what node this is! Grr...")
}

func (c *CTemplateSet) compileIdentSwitchN(con CContext, n *parse.CommandNode) (out string) {
	c.detail("in compileIdentSwitchN")
	out, _, _, _ = c.compileIdentSwitch(con, n)
	return out
}

func (c *CTemplateSet) dumpSymbol(pos int, n *parse.CommandNode, symbol string) {
	c.detail("symbol:", symbol)
	c.detail("n.Args[pos+1]", n.Args[pos+1])
	c.detail("n.Args[pos+2]", n.Args[pos+2])
}

func (c *CTemplateSet) compareFunc(con CContext, pos int, n *parse.CommandNode, compare string) (out string) {
	c.dumpSymbol(pos, n, compare)
	return c.compileIfVarSubN(con, n.Args[pos+1].String()) + " " + compare + " " + c.compileIfVarSubN(con, n.Args[pos+2].String())
}

func (c *CTemplateSet) simpleMath(con CContext, pos int, n *parse.CommandNode, symbol string) (out string, val reflect.Value) {
	leftParam, val2 := c.compileIfVarSub(con, n.Args[pos+1].String())
	rightParam, val3 := c.compileIfVarSub(con, n.Args[pos+2].String())
	if val2.IsValid() {
		val = val2
	} else if val3.IsValid() {
		val = val3
	} else {
		// TODO: What does this do?
		numSample := 1
		val = reflect.ValueOf(numSample)
	}
	c.dumpSymbol(pos, n, symbol)
	return leftParam + " " + symbol + " " + rightParam, val
}

func (c *CTemplateSet) compareJoin(con CContext, pos int, node *parse.CommandNode, symbol string) (pos2 int, out string) {
	c.detailf("Building %s function", symbol)
	if pos == 0 {
		c.logger.Println("pos:", pos)
		panic(symbol + " is missing a left operand")
	}
	if len(node.Args) <= pos {
		c.logger.Println("post pos:", pos)
		c.logger.Println("len(node.Args):", len(node.Args))
		panic(symbol + " is missing a right operand")
	}

	left := c.compileBoolSub(con, node.Args[pos-1].String())
	_, funcExists := c.funcMap[node.Args[pos+1].String()]

	var right string
	if !funcExists {
		right = c.compileBoolSub(con, node.Args[pos+1].String())
	}
	out = left + " " + symbol + " " + right

	c.detail("Left op:", node.Args[pos-1])
	c.detail("Right op:", node.Args[pos+1])
	if !funcExists {
		pos++
	}
	c.detail("pos:", pos)
	c.detail("len(node.Args):", len(node.Args))

	return pos, out
}

func (c *CTemplateSet) compileIdentSwitch(con CContext, node *parse.CommandNode) (out string, val reflect.Value, literal, notIdent bool) {
	c.dumpCall("compileIdentSwitch", con, node)
	litString := func(inner string, bytes bool) {
		if !bytes {
			inner = "StringToBytes(" + inner + "/*,tmp*/)"
		}
		out = "w.Write(" + inner + ")\n"
		literal = true
	}
ArgLoop:
	for pos := 0; pos < len(node.Args); pos++ {
		id := node.Args[pos]
		c.detail("pos:", pos)
		c.detail("id:", id)
		switch id.String() {
		case "not":
			out += "!"
		case "or", "and":
			var rout string
			pos, rout = c.compareJoin(con, pos, node, c.funcMap[id.String()].(string)) // TODO: Test this
			out += rout
		case "le", "lt", "gt", "ge":
			out += c.compareFunc(con, pos, node, c.funcMap[id.String()].(string))
			break ArgLoop
		case "eq", "ne":
			o := c.compareFunc(con, pos, node, c.funcMap[id.String()].(string))
			if out == "!" {
				o = "(" + o + ")"
			}
			out += o
			break ArgLoop
		case "add", "subtract", "divide", "multiply":
			rout, rval := c.simpleMath(con, pos, node, c.funcMap[id.String()].(string))
			out += rout
			val = rval
			break ArgLoop
		case "elapsed":
			leftOp := node.Args[pos+1].String()
			leftParam, _ := c.compileIfVarSub(con, leftOp)
			// TODO: Refactor this
			// TODO: Validate that this is actually a time.Time
			//litString("time.Since("+leftParam+").String()", false)
			c.importMap["time"] = "time"
			c.importMap["github.com/Azareal/Gosora/uutils"] = "github.com/Azareal/Gosora/uutils"
			litString("time.Duration(uutils.Nanotime() - "+leftParam+").String()", false)
			break ArgLoop
		case "dock":
			// TODO: Implement string literals properly
			leftOp := node.Args[pos+1].String()
			rightOp := node.Args[pos+2].String()
			if len(leftOp) == 0 || len(rightOp) == 0 {
				panic("The left or right operand for function dock cannot be left blank")
			}
			leftParam := leftOp
			if leftOp[0] != '"' {
				leftParam, _ = c.compileIfVarSub(con, leftParam)
			}
			if rightOp[0] == '"' {
				panic("The right operand for function dock cannot be a string")
			}
			rightParam, val3 := c.compileIfVarSub(con, rightOp)
			if !val3.IsValid() {
				panic("val3 is invalid")
			}
			val = val3

			// TODO: Refactor this
			if leftParam[0] == '"' {
				leftParam = strings.TrimSuffix(strings.TrimPrefix(leftParam, "\""), "\"")
				id, ok := c.config.DockToID[leftParam]
				if ok {
					out = "c.BuildWidget3(" + strconv.Itoa(id) + "," + rightParam + ")\n"
					literal = true
					break ArgLoop
				}
			}
			litString("c.BuildWidget("+leftParam+","+rightParam+")", false)
			break ArgLoop
		case "hasWidgets":
			// TODO: Implement string literals properly
			leftOp := node.Args[pos+1].String()
			rightOp := node.Args[pos+2].String()
			if len(leftOp) == 0 || len(rightOp) == 0 {
				panic("The left or right operand for function dock cannot be left blank")
			}
			leftParam := leftOp
			if leftOp[0] != '"' {
				leftParam, _ = c.compileIfVarSub(con, leftParam)
			}
			if rightOp[0] == '"' {
				panic("The right operand for function dock cannot be a string")
			}
			rightParam, val3 := c.compileIfVarSub(con, rightOp)
			if !val3.IsValid() {
				panic("val3 is invalid")
			}
			val = val3

			// TODO: Refactor this
			if leftParam[0] == '"' {
				leftParam = strings.TrimSuffix(strings.TrimPrefix(leftParam, "\""), "\"")
				id, ok := c.config.DockToID[leftParam]
				if ok {
					out = "c.HasWidgets2(" + strconv.Itoa(id) + "," + rightParam + ")"
					literal = true
					break ArgLoop
				}
			}
			out = "c.HasWidgets(" + leftParam + "," + rightParam + ")"
			literal = true
			break ArgLoop
		case "js":
			if c.lang == "js" {
				out = "true"
			} else {
				out = "false"
			}
			literal = true
			break ArgLoop
		case "lang":
			// TODO: Implement string literals properly
			leftOp := node.Args[pos+1].String()
			if len(leftOp) == 0 {
				panic("The left operand for the language string cannot be left blank")
			}
			if leftOp[0] == '"' {
				// ! Slightly crude but it does the job
				leftParam := strings.Replace(leftOp, "\"", "", -1)
				c.langIndexToName = append(c.langIndexToName, leftParam)
				notIdent = true
				con.PushPhrase(len(c.langIndexToName) - 1)
			} else {
				leftParam := leftOp
				if leftOp[0] != '"' {
					leftParam, _ = c.compileIfVarSub(con, leftParam)
				}
				// TODO: Add an optimisation if it's a string literal passsed in from a parent template rather than a true dynamic
				litString("phrases.GetTmplPhrasef("+leftParam+")", false)
				c.importMap[langPkg] = langPkg
			}
			break ArgLoop
		case "langf":
			// TODO: Implement string literals properly
			leftOp := node.Args[pos+1].String()
			if len(leftOp) == 0 {
				panic("The left operand for the language string cannot be left blank")
			}
			if leftOp[0] != '"' {
				panic("Phrase names cannot be dynamic")
			}

			var olist []string
			for i := pos + 2; i < len(node.Args); i++ {
				op := node.Args[i].String()
				if op != "" {
					if /*op[0] == '.' || */ op[0] == '$' {
						panic("langf args cannot be dynamic")
					}
					if op[0] != '.' && op[0] != '"' && !unicode.IsDigit(rune(op[0])) {
						break
					}
					olist = append(olist, op)
				}
			}
			if len(olist) == 0 {
				panic("You must provide parameters for langf")
			}

			ob := ","
			for _, op := range olist {
				if op[0] == '.' {
					param, val3 := c.compileIfVarSub(con, op)
					if !val3.IsValid() {
						panic("val3 is invalid")
					}
					ob += param + ","
					continue
				}
				allNum := true
				for _, o := range op {
					if !unicode.IsDigit(o) {
						allNum = false
					}
				}
				if allNum {
					ob += strings.Replace(op, "\"", "\\\"", -1) + ","
				} else {
					ob += ob + ","
				}
			}
			if ob != "" {
				ob = ob[:len(ob)-1]
			}

			// TODO: Implement string literals properly
			// ! Slightly crude but it does the job
			litString("phrases.GetTmplPhrasef("+leftOp+ob+")", false)
			c.importMap[langPkg] = langPkg
			break ArgLoop
		case "level":
			// TODO: Implement level literals
			leftOp := node.Args[pos+1].String()
			if len(leftOp) == 0 {
				panic("The leftoperand for function level cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOp)
			// TODO: Refactor this
			litString("phrases.GetLevelPhrase("+leftParam+")", false)
			c.importMap[langPkg] = langPkg
			break ArgLoop
		case "bunit":
			// TODO: Implement bunit literals
			leftOp := node.Args[pos+1].String()
			if len(leftOp) == 0 {
				panic("The leftoperand for function buint cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOp)
			out = "{\nbyteFloat, unit := c.ConvertByteUnit(float64(" + leftParam + "))\n"
			out += "w.Write(StringToBytes(fmt.Sprintf(\"%.1f\", byteFloat)/*,tmp*/))\nw.Write(StringToBytes(unit/*,tmp*/))\n}\n"
			literal = true
			c.importMap["fmt"] = "fmt"
			break ArgLoop
		case "abstime":
			// TODO: Implement level literals
			leftOp := node.Args[pos+1].String()
			if len(leftOp) == 0 {
				panic("The leftoperand for function abstime cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOp)
			// TODO: Refactor this
			litString(leftParam+".Format(\"2006-01-02 15:04:05\")", false)
			break ArgLoop
		case "reltime":
			// TODO: Implement level literals
			leftOp := node.Args[pos+1].String()
			if len(leftOp) == 0 {
				panic("The leftoperand for function reltime cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOp)
			// TODO: Refactor this
			litString("c.RelativeTime("+leftParam+")", false)
			break ArgLoop
		case "scope":
			literal = true
			break ArgLoop
		// TODO: Optimise ptmpl
		case "dyntmpl", "ptmpl":
			var pageParam, headParam string
			// TODO: Implement string literals properly
			// TODO: Should we check to see if pos+3 is within the bounds of the slice?
			nameOp := node.Args[pos+1].String()
			pageOp := node.Args[pos+2].String()
			headOp := node.Args[pos+3].String()
			if len(nameOp) == 0 || len(pageOp) == 0 || len(headOp) == 0 {
				panic("None of the three operands for function dyntmpl can be left blank")
			}
			nameParam := nameOp
			if nameOp[0] != '"' {
				nameParam, _ = c.compileIfVarSub(con, nameParam)
			}
			if pageOp[0] == '"' {
				panic("The page operand for function dyntmpl cannot be a string")
			}
			if headOp[0] == '"' {
				panic("The head operand for function dyntmpl cannot be a string")
			}

			pageParam, val3 := c.compileIfVarSub(con, pageOp)
			if !val3.IsValid() {
				panic("val3 is invalid")
			}
			headParam, val4 := c.compileIfVarSub(con, headOp)
			if !val4.IsValid() {
				panic("val4 is invalid")
			}
			val = val4

			// TODO: Refactor this
			// TODO: Call the template function directly rather than going through RunThemeTemplate to eliminate a round of indirection?
			out = "{\ne := " + headParam + ".Theme.RunTmpl(" + nameParam + "," + pageParam + ",w)\n"
			out += "if e != nil {\nreturn e\n}\n}\n"
			literal = true
			break ArgLoop
		case "flush":
			literal = true
			break ArgLoop
		/*if c.lang == "js" {
			continue
		}
		out = "if fl, ok := iw.(http.Flusher); ok {\nfl.Flush()\n}\n"
		literal = true
		c.importMap["net/http"] = "net/http"
		break ArgLoop*/
		// TODO: Test this
		case "res":
			leftOp := node.Args[pos+1].String()
			if len(leftOp) == 0 {
				panic("The leftoperand for function res cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOp)
			literal = true
			if leftParam[0] == '"' {
				if leftParam[1] == '/' && leftParam[2] == '/' {
					litString(leftParam, false)
					break ArgLoop
				}
				out = "{n := " + leftParam + "\nif f, ok := c.StaticFiles.GetShort(n); ok {\nw.Write(StringToBytes(f.OName))\n} else {\nw.Write(StringToBytes(n))\n}}\n"
				break ArgLoop
			}
			out = "{n := " + leftParam + "\nif n[0] == '/' && n[1] == '/' {\n} else {\nif f, ok := c.StaticFiles.GetShort(n); ok {\nn = f.OName\n}\nw.Write(StringToBytes(n))\n}\n"
			break ArgLoop
		default:
			c.detail("Variable!")
			if len(node.Args) > (pos + 1) {
				nextNode := node.Args[pos+1].String()
				if nextNode == "or" || nextNode == "and" {
					continue
				}
			}
			out += c.compileIfVarSubN(con, id.String())
		}
	}
	c.retCall("compileIdentSwitch", out, val, literal)
	return out, val, literal, notIdent
}

func (c *CTemplateSet) compileReflectSwitch(con CContext, node *parse.CommandNode) (out string, outVal reflect.Value) {
	c.dumpCall("compileReflectSwitch", con, node)
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
	case *parse.FieldNode:
		if c.config.SuperDebug {
			c.logger.Println("Field Node:", n.Ident)
			for _, id := range n.Ident {
				c.logger.Println("Field Bit:", id)
			}
		}
		/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Coming Soon. */
		return c.compileIfVarSub(con, n.String())
	case *parse.ChainNode:
		c.detail("Chain Node:", n.Node)
		c.detail("node.Args:", node.Args)
	case *parse.DotNode:
		return con.VarHolder, con.HoldReflect
	case *parse.NilNode:
		panic("Nil is not a command x.x")
	default:
		//panic("I don't know what node this is")
	}
	return out, outVal
}

func (c *CTemplateSet) compileIfVarSubN(con CContext, varname string) (out string) {
	c.dumpCall("compileIfVarSubN", con, varname)
	out, _ = c.compileIfVarSub(con, varname)
	return out
}

func (c *CTemplateSet) compileIfVarSub(con CContext, varname string) (out string, val reflect.Value) {
	c.dumpCall("compileIfVarSub", con, varname)
	cur := con.HoldReflect
	if varname[0] != '.' && varname[0] != '$' {
		return varname, cur
	}

	stepInterface := func() {
		nobreak := (cur.Type().Name() == "nobreak")
		c.detailf("cur.Type().Name(): %+v\n", cur.Type().Name())
		if cur.Kind() == reflect.Interface && !nobreak {
			cur = cur.Elem()
			out += ".(" + cur.Type().Name() + ")"
		}
	}

	bits := strings.Split(varname, ".")
	if varname[0] == '$' {
		var res VarItemReflect
		if varname[1] == '.' {
			res = c.localVars[con.TemplateName]["."]
		} else {
			res = c.localVars[con.TemplateName][strings.TrimPrefix(bits[0], "$")]
		}
		out += res.Destination
		cur = res.Value

		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
		}
	} else {
		out += con.VarHolder
		stepInterface()
	}
	bits[0] = strings.TrimPrefix(bits[0], "$")

	dumpKind := func(pre string) {
		c.detail(pre+" Kind:", cur.Kind())
		c.detail(pre+" Type:", cur.Type().Name())
	}
	dumpKind("Cur")
	for _, bit := range bits {
		c.detail("Variable Field:", bit)
		if bit == "" {
			continue
		}

		// TODO: Fix this up so that it works for regular pointers and not just struct pointers. Ditto for the other cur.Kind() == reflect.Ptr we have in this file
		if cur.Kind() == reflect.Ptr {
			c.detail("Looping over pointer")
			for cur.Kind() == reflect.Ptr {
				cur = cur.Elem()
			}
			c.detail("Data Kind:", cur.Kind().String())
			c.detail("Field Bit:", bit)
		}

		cur = cur.FieldByName(bit)
		out += "." + bit
		if !cur.IsValid() {
			c.logger.Println("cur: ", cur)
			panic(out + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		stepInterface()
		if !cur.IsValid() {
			c.logger.Println("cur: ", cur)
			panic(out + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		dumpKind("Data")
	}

	c.detail("Out Value:", out)
	dumpKind("Out")
	for _, varItem := range c.varList {
		if strings.HasPrefix(out, varItem.Destination) {
			out = strings.Replace(out, varItem.Destination, varItem.Name, 1)
		}
	}

	_, ok := c.stats[out]
	if ok {
		c.stats[out]++
	} else {
		c.stats[out] = 1
	}

	c.retCall("compileIfVarSub", out, cur)
	return out, cur
}

func (c *CTemplateSet) compileBoolSub(con CContext, varname string) string {
	c.dumpCall("compileBoolSub", con, varname)
	out, val := c.compileIfVarSub(con, varname)
	// TODO: What if it's a pointer or an interface? I *think* we've got pointers handled somewhere, but not interfaces which we don't know the types of at compile time
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		out += ">0"
	case reflect.Bool: // Do nothing
	case reflect.String:
		out += "!=\"\""
	case reflect.Slice, reflect.Map:
		out = "len(" + out + ")!=0"
	// TODO: Follow the pointer and evaluate it?
	case reflect.Ptr:
		out += "!=nil"
	default:
		c.logger.Println("Variable Name:", varname)
		c.logger.Println("Variable Holder:", con.VarHolder)
		c.logger.Println("Variable Kind:", con.HoldReflect.Kind())
		panic("I don't know what this variable's type is o.o\n")
	}
	c.retCall("compileBoolSub", out)
	return out
}

// For debugging the template generator
func (c *CTemplateSet) debugParam(param interface{}, depth int) (pstr string) {
	switch p := param.(type) {
	case CContext:
		return "con,"
	case reflect.Value:
		if p.Kind() == reflect.Ptr || p.Kind() == reflect.Interface {
			for p.Kind() == reflect.Ptr || p.Kind() == reflect.Interface {
				if p.Kind() == reflect.Ptr {
					pstr += "*"
				} else {
					pstr += "£"
				}
				p = p.Elem()
			}
		}
		kind := p.Kind().String()
		if kind != "struct" {
			pstr += kind
		} else {
			pstr += p.Type().Name()
		}
		return pstr + ","
	case string:
		return "\"" + p + "\","
	case int:
		return strconv.Itoa(p) + ","
	case bool:
		if p {
			return "true,"
		}
		return "false,"
	case func(string) string:
		if p == nil {
			return "nil,"
		}
		return "func(string) string),"
	default:
		return "?,"
	}
}
func (c *CTemplateSet) dumpCall(name string, params ...interface{}) {
	var pstr string
	for _, param := range params {
		pstr += c.debugParam(param, 0)
	}
	if len(pstr) > 0 {
		pstr = pstr[:len(pstr)-1]
	}
	c.detail("called " + name + "(" + pstr + ")")
}
func (c *CTemplateSet) retCall(name string, params ...interface{}) {
	var pstr string
	for _, param := range params {
		pstr += c.debugParam(param, 0)
	}
	if len(pstr) > 0 {
		pstr = pstr[:len(pstr)-1]
	}
	c.detail("returned from " + name + " => (" + pstr + ")")
}

func buildUserExprs(holder string) ([]string, []string) {
	userExprs := []string{
		holder + ".CurrentUser.Loggedin",
		holder + ".CurrentUser.IsSuperMod",
		holder + ".CurrentUser.IsAdmin",
	}
	negUserExprs := []string{
		"!" + holder + ".CurrentUser.Loggedin",
		"!" + holder + ".CurrentUser.IsSuperMod",
		"!" + holder + ".CurrentUser.IsAdmin",
	}
	return userExprs, negUserExprs
}

func (c *CTemplateSet) compileVarSub(con CContext, varname string, val reflect.Value, assLines string, onEnd func(string) string) {
	c.dumpCall("compileVarSub", con, varname, val, assLines, onEnd)
	defer c.retCall("compileVarSub")
	if onEnd == nil {
		onEnd = func(in string) string {
			return in
		}
	}

	// Is this a literal string?
	if len(varname) != 0 && varname[0] == '"' {
		con.Push("lvarsub", onEnd(assLines+"w.Write(StringToBytes("+varname+"/*,tmp*/))\n"))
		return
	}
	for _, varItem := range c.varList {
		if strings.HasPrefix(varname, varItem.Destination) {
			varname = strings.Replace(varname, varItem.Destination, varItem.Name, 1)
		}
	}

	_, ok := c.stats[varname]
	if ok {
		c.stats[varname]++
	} else {
		c.stats[varname] = 1
	}
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() == reflect.Ptr {
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
			varname = "*" + varname
		}
	}

	c.detail("varname:", varname)
	c.detail("assLines:", assLines)
	var base string
	switch val.Kind() {
	case reflect.Int:
		c.importMap["strconv"] = "strconv"
		base = "StringToBytes(strconv.Itoa(" + varname + ")/*,tmp*/)"
	case reflect.Bool:
		// TODO: Take c.memberOnly into account
		// TODO: Make this a template fragment so more optimisations can be applied to this
		// TODO: De-duplicate this logic
		userExprs, negUserExprs := buildUserExprs(con.RootHolder)
		if c.guestOnly {
			c.detail("optimising away member branch")
			if inSlice(userExprs, varname) {
				c.detail("positive condition:", varname)
				c.addText(con, []byte("false"))
				return
			} else if inSlice(negUserExprs, varname) {
				c.detail("negative condition:", varname)
				c.addText(con, []byte("true"))
				return
			}
		} else if c.memberOnly {
			c.detail("optimising away guest branch")
			if (con.RootHolder + ".CurrentUser.Loggedin") == varname {
				c.detail("positive condition:", varname)
				c.addText(con, []byte("true"))
				return
			} else if ("!" + con.RootHolder + ".CurrentUser.Loggedin") == varname {
				c.detail("negative condition:", varname)
				c.addText(con, []byte("false"))
				return
			}
		}
		startIf := con.StartIf("if " + varname + " {\n")
		c.addText(con, []byte("true"))
		con.EndIf(startIf, "} ")
		con.Push("startelse", "else {\n")
		c.addText(con, []byte("false"))
		con.Push("endelse", "}\n")
		return
	case reflect.Slice:
		if val.Len() == 0 {
			c.critical("varname:", varname)
			panic("The sample data needs at-least one or more elements for the slices. We're looking into removing this requirement at some point!")
		}
		item := val.Index(0)
		if item.Type().Name() != "uint8" { // uint8 == byte, complicated because it's a type alias
			panic("unable to format " + item.Type().Name() + " as text")
		}
		base = varname
	case reflect.String:
		if val.Type().Name() != "string" && !strings.HasPrefix(varname, "string(") {
			varname = "string(" + varname + ")"
		}
		base = "StringToBytes(" + varname + "/*,tmp*/)"
		// We don't to waste time on this conversion / w.Write call when guests don't have sessions
		// TODO: Implement this properly
		if c.guestOnly && base == "StringToBytes("+con.RootHolder+".CurrentUser.Session/*,tmp*/))" {
			return
		}
	case reflect.Int8, reflect.Int16, reflect.Int32:
		c.importMap["strconv"] = "strconv"
		base = "StringToBytes(strconv.FormatInt(int64(" + varname + "), 10)/*,tmp*/)"
	case reflect.Int64:
		c.importMap["strconv"] = "strconv"
		base = "StringToBytes(strconv.FormatInt(" + varname + ", 10)/*,tmp*/)"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		c.importMap["strconv"] = "strconv"
		base = "StringToBytes(strconv.FormatUint(uint64(" + varname + "), 10)/*,tmp*/)"
	case reflect.Uint64:
		c.importMap["strconv"] = "strconv"
		base = "StringToBytes(strconv.FormatUint(" + varname + ", 10)/*,tmp*/)"
	case reflect.Struct:
		// TODO: Avoid clashing with other packages which have structs named Time
		if val.Type().Name() == "Time" {
			base = "StringToBytes(" + varname + ".String()/*,tmp*/)"
		} else {
			if !val.IsValid() {
				panic(assLines + varname + "^\n" + "Invalid value. Maybe, it doesn't exist?")
			}
			c.logger.Println("Unknown Struct Name:", varname)
			c.logger.Println("Unknown Struct:", val.Type().Name())
			panic("-- I don't know what this variable's type is o.o\n")
		}
	default:
		if !val.IsValid() {
			panic(assLines + varname + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		c.logger.Println("Unknown Variable Name:", varname)
		c.logger.Println("Unknown Kind:", val.Kind())
		c.logger.Println("Unknown Type:", val.Type().Name())
		panic("-- I don't know what this variable's type is o.o\n")
	}
	c.detail("base:", base)
	if assLines == "" {
		con.Push("varsub", base)
	} else {
		con.Push("lvarsub", onEnd(assLines+base))
	}
}

func (c *CTemplateSet) compileSubTemplate(pcon CContext, node *parse.TemplateNode) {
	c.dumpCall("compileSubTemplate", pcon, node)
	defer c.retCall("compileSubTemplate")
	c.detail("Template Node: ", node.Name)

	fname := strings.TrimSuffix(node.Name, filepath.Ext(node.Name))
	if c.themeName != "" {
		_, ok := c.perThemeTmpls[fname]
		if !ok {
			c.detail("fname not in c.perThemeTmpls")
			c.detail("c.perThemeTmpls", c.perThemeTmpls)
		}
		fname += "_" + c.themeName
	}
	if c.guestOnly {
		fname += "_guest"
	} else if c.memberOnly {
		fname += "_member"
	}

	_, ok := c.templateList[fname]
	if !ok {
		// TODO: Cascade errors back up the tree to the caller?
		content, err := c.loadTemplate(c.fileDir, node.Name)
		if err != nil {
			c.logger.Fatal(err)
		}

		//tree := parse.New(node.Name, c.funcMap)
		//treeSet := make(map[string]*parse.Tree)
		treeSet, err := parse.Parse(node.Name, content, "{{", "}}", c.funcMap)
		if err != nil {
			c.logger.Fatal(err)
		}
		c.detailf("treeSet: %+v\n", treeSet)

		for nname, tree := range treeSet {
			if node.Name == nname {
				c.templateList[fname] = tree
			} else {
				if !strings.HasPrefix(nname, ".html") {
					c.templateList[nname] = tree
				} else {
					c.templateList[strings.TrimSuffix(nname, ".html")] = tree
				}
			}
		}
		c.detailf("c.templateList: %+v\n", c.templateList)
	}

	con := pcon
	con.VarHolder = "tmpl_" + fname + "_vars"
	con.TemplateName = fname
	if node.Pipe != nil {
		for _, cmd := range node.Pipe.Cmds {
			switch p := cmd.Args[0].(type) {
			case *parse.FieldNode:
				// TODO: Incomplete but it should cover the basics
				cur := pcon.HoldReflect
				var varBit string
				if cur.Kind() == reflect.Interface {
					cur = cur.Elem()
					varBit += ".(" + cur.Type().Name() + ")"
				}

				for _, id := range p.Ident {
					c.detail("Data Kind:", cur.Kind().String())
					c.detail("Field Bit:", id)
					cur = c.skipStructPointers(cur, id)
					c.checkIfValid(cur, pcon.VarHolder, pcon.HoldReflect, varBit, false)

					if cur.Kind() != reflect.Interface {
						cur = cur.FieldByName(id)
						varBit += "." + id
					}

					// TODO: Handle deeply nested pointers mixed with interfaces mixed with pointers better
					if cur.Kind() == reflect.Interface {
						cur = cur.Elem()
						varBit += ".("
						// TODO: Surely, there's a better way of doing this?
						if cur.Type().PkgPath() != "main" && cur.Type().PkgPath() != "" {
							c.importMap["html/template"] = "html/template"
							varBit += strings.TrimPrefix(cur.Type().PkgPath(), "html/") + "."
						}
						varBit += cur.Type().Name() + ")"
					}
				}
				con.VarHolder = pcon.VarHolder + varBit
				con.HoldReflect = cur
			case *parse.StringNode:
				//con.VarHolder = pcon.VarHolder
				//con.HoldReflect = pcon.HoldReflect
				con.VarHolder = p.Quoted
				con.HoldReflect = reflect.ValueOf(p.Quoted)
			case *parse.DotNode:
				con.VarHolder = pcon.VarHolder
				con.HoldReflect = pcon.HoldReflect
			case *parse.NilNode:
				panic("Nil is not a command x.x")
			default:
				c.critical("Unknown Param Type:", p)
				pvar := reflect.ValueOf(p)
				c.critical("param kind:", pvar.Kind().String())
				c.critical("param type:", pvar.Type().Name())
				if pvar.Kind() == reflect.Ptr {
					c.critical("Looping over pointer")
					for pvar.Kind() == reflect.Ptr {
						pvar = pvar.Elem()
					}
					c.critical("concrete kind:", pvar.Kind().String())
					c.critical("concrete type:", pvar.Type().Name())
				}
				panic("")
			}
		}
	}

	//c.templateList[fname] = tree
	subtree := c.templateList[fname]
	c.detail("subtree.Root", subtree.Root)
	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".", con.VarHolder, con.HoldReflect}
	c.fragmentCursor[fname] = 0

	var startBit, endBit string
	if con.LoopDepth != 0 {
		startBit = "{\n"
		endBit = "}\n"
	}
	con.StartTemplate(startBit)
	c.rootIterate(subtree, con)
	con.EndTemplate(endBit)
	//c.templateFragmentCount[fname] = c.fragmentCursor[fname] + 1
	if _, ok := c.fragOnce[fname]; !ok {
		c.fragOnce[fname] = true
	}

	// map[string]map[string]bool
	c.detail("overridenTrack loop")
	c.detail("fname:", fname)
	for themeName, track := range c.overridenTrack {
		c.detail("themeName:", themeName)
		c.detailf("track: %+v\n", track)
		croot, ok := c.overridenRoots[themeName]
		if !ok {
			croot = make(map[string]bool)
			c.overridenRoots[themeName] = croot
		}
		c.detailf("croot: %+v\n", croot)
		for tmplName, _ := range track {
			cname := tmplName
			if c.guestOnly {
				cname += "_guest"
			} else if c.memberOnly {
				cname += "_member"
			}
			c.detail("cname:", cname)
			if fname == cname {
				c.detail("match")
				croot[strings.TrimSuffix(strings.TrimSuffix(con.RootTemplateName, "_guest"), "_member")] = true
			} else {
				c.detail("no match")
			}
		}
	}
	c.detailf("c.overridenRoots: %+v\n", c.overridenRoots)
}

func (c *CTemplateSet) loadTemplate(fileDir, name string) (content string, err error) {
	c.dumpCall("loadTemplate", fileDir, name)
	c.detail("c.themeName:", c.themeName)
	if c.themeName != "" {
		t := "./themes/" + c.themeName + "/overrides/" + name
		c.detail("per-theme override:", true)
		res, err := ioutil.ReadFile(t)
		if err == nil {
			content = string(res)
			if c.config.Minify {
				content = Minify(content)
			}
			return content, nil
		}
		c.detail("override err:", err)
	}

	res, err := ioutil.ReadFile(c.fileDir + "overrides/" + name)
	if err != nil {
		c.detail("override path:", c.fileDir+"overrides/"+name)
		c.detail("override err:", err)
		res, err = ioutil.ReadFile(c.fileDir + name)
		if err != nil {
			return "", err
		}
	}
	content = string(res)
	if c.config.Minify {
		content = Minify(content)
	}
	return content, nil
}

func (c *CTemplateSet) afterTemplate(con CContext, startIndex int /*, svmap map[string]int*/) {
	c.dumpCall("afterTemplate", con, startIndex)
	defer c.retCall("afterTemplate")

	loopDepth := 0
	ifNilDepth := 0
	var outBuf = *con.OutBuf
	varcounts := make(map[string]int)
	loopStart := startIndex
	otype := outBuf[startIndex].Type
	if otype == "startloop" && (len(outBuf) > startIndex+1) {
		loopStart++
	}
	if otype == "startif" && (len(outBuf) > startIndex+1) {
		loopStart++
	}

	// Exclude varsubs within loops for now
OLoop:
	for i := loopStart; i < len(outBuf); i++ {
		item := outBuf[i]
		c.detail("item:", item)
		switch item.Type {
		case "startloop":
			loopDepth++
			c.detail("loopDepth:", loopDepth)
		case "endloop":
			loopDepth--
			c.detail("loopDepth:", loopDepth)
			if loopDepth == -1 {
				break OLoop
			}
		case "startif":
			if item.Extra.(bool) == true {
				ifNilDepth++
			}
		case "endif":
			item2 := outBuf[item.Extra.(int)]
			if item2.Extra.(bool) == true {
				ifNilDepth--
			}
			if ifNilDepth == -1 {
				break OLoop
			}
		case "varsub":
			if loopDepth == 0 && ifNilDepth == 0 {
				count := varcounts[item.Body]
				varcounts[item.Body] = count + 1
				c.detail("count " + strconv.Itoa(count) + " for " + item.Body)
				c.detail("loopDepth:", loopDepth)
			}
		}
	}

	var varstr string
	var i int
	varmap := make(map[string]int)
	/*for svkey, sventry := range svmap {
		varmap[svkey] = sventry
	}*/
	for name, count := range varcounts {
		if count > 1 {
			varstr += "var c_v_" + strconv.Itoa(i) + "=" + name + "\n"
			varmap[name] = i
			i++
		}
	}

	// Exclude varsubs within loops for now
	loopDepth = 0
	ifNilDepth = 0
OOLoop:
	for i := loopStart; i < len(outBuf); i++ {
		item := outBuf[i]
		switch item.Type {
		case "startloop":
			loopDepth++
		case "endloop":
			loopDepth--
			if loopDepth == -1 {
				break OOLoop
			} //con.Push("startif", "if "+varname+" {\n")
		case "startif":
			if item.Extra.(bool) == true {
				ifNilDepth++
			}
		case "endif":
			item2 := outBuf[item.Extra.(int)]
			if item2.Extra.(bool) == true {
				ifNilDepth--
			}
			if ifNilDepth == -1 {
				break OOLoop
			}
		case "varsub":
			if loopDepth == 0 && ifNilDepth == 0 {
				index, ok := varmap[item.Body]
				if ok {
					item.Body = "c_v_" + strconv.Itoa(index)
					item.Type = "cvarsub"
					outBuf[i] = item
				}
			}
		}
	}

	con.AttachVars(varstr, startIndex)
}

const (
	ATTmpl = iota
	ATLoop
	ATIfPtr
)

func (c *CTemplateSet) afterTemplateV2(con CContext, startIndex int /*, typ int*/, svmap map[string]int) {
	c.dumpCall("afterTemplateV2", con, startIndex)
	defer c.retCall("afterTemplateV2")

	loopDepth, ifNilDepth := 0, 0
	var outBuf = *con.OutBuf
	varcounts := make(map[string]int)
	loopStart := startIndex
	otype := outBuf[startIndex].Type
	if otype == "startloop" && (len(outBuf) > startIndex+1) {
		loopStart++
	}
	if otype == "startif" && (len(outBuf) > startIndex+1) {
		loopStart++
	}

	// Exclude varsubs within loops for now
OLoop:
	for i := loopStart; i < len(outBuf); i++ {
		item := outBuf[i]
		c.detail("item:", item)
		switch item.Type {
		case "startloop":
			loopDepth++
			c.detail("loopDepth:", loopDepth)
		case "endloop":
			loopDepth--
			c.detail("loopDepth:", loopDepth)
			if loopDepth == -1 {
				break OLoop
			}
		case "startif":
			if item.Extra.(bool) == true {
				ifNilDepth++
			}
		case "endif":
			item2 := outBuf[item.Extra.(int)]
			if item2.Extra.(bool) == true {
				ifNilDepth--
			}
			if ifNilDepth == -1 {
				break OLoop
			}
		case "varsub":
			if loopDepth == 0 && ifNilDepth == 0 {
				count := varcounts[item.Body]
				varcounts[item.Body] = count + 1
				c.detail("count " + strconv.Itoa(count) + " for " + item.Body)
				c.detail("loopDepth:", loopDepth)
			}
		}
	}

	var varstr string
	var i int
	varmap := make(map[string]int)
	/*for svkey, sventry := range svmap {
		varmap[svkey] = sventry
	}*/
	for name, count := range varcounts {
		if count > 1 {
			varstr += "var c_v_" + strconv.Itoa(i) + "=" + name + "\n"
			varmap[name] = i
			i++
		}
	}

	// Exclude varsubs within loops for now
	loopDepth, ifNilDepth = 0, 0
OOLoop:
	for i := loopStart; i < len(outBuf); i++ {
		item := outBuf[i]
		switch item.Type {
		case "startloop":
			loopDepth++
		case "endloop":
			loopDepth--
			if loopDepth == -1 {
				break OOLoop
			} //con.Push("startif", "if "+varname+" {\n")
		case "startif":
			if item.Extra.(bool) == true {
				ifNilDepth++
			}
		case "endif":
			item2 := outBuf[item.Extra.(int)]
			if item2.Extra.(bool) == true {
				ifNilDepth--
			}
			if ifNilDepth == -1 {
				break OOLoop
			}
		case "varsub":
			if loopDepth == 0 && ifNilDepth == 0 {
				index, ok := varmap[item.Body]
				if ok {
					item.Body = "c_v_" + strconv.Itoa(index)
					item.Type = "cvarsub"
					outBuf[i] = item
				}
			}
		}
	}

	con.AttachVars(varstr, startIndex)
}

// TODO: Should we rethink the way the log methods work or their names?

func (c *CTemplateSet) detail(args ...interface{}) {
	if c.config.SuperDebug {
		c.logger.Println(args...)
	}
}

func (c *CTemplateSet) detailf(left string, args ...interface{}) {
	if c.config.SuperDebug {
		c.logger.Printf(left, args...)
	}
}

func (c *CTemplateSet) error(args ...interface{}) {
	if c.config.Debug {
		c.logger.Println(args...)
	}
}

func (c *CTemplateSet) critical(args ...interface{}) {
	c.logger.Println(args...)
}
