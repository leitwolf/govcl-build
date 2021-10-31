package main

import (
	"bytes"
	"embed"
	"flag"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/hallazzang/syso"
	"github.com/hallazzang/syso/pkg/coff"
)

//go:embed files
var embedFiles embed.FS

var (
	icon        string
	outName     string
	showConsole bool
	embedDll    bool
	reduceSize  bool
)

func main() {
	initConfigs()
	genIcon()
	copyManifest()
	createSyso()
	build()
	clear()
}

func initConfigs() {
	flag.StringVar(&icon, "icon", "", "图标(png)")
	flag.StringVar(&icon, "i", "", "图标(png) [缩写]")
	flag.StringVar(&outName, "name", "", "生成文件名")
	flag.StringVar(&outName, "n", "", "生成文件名 [缩写]")
	flag.BoolVar(&showConsole, "showConsole", false, "显示console窗口")
	flag.BoolVar(&showConsole, "s", false, "显示console窗口 [缩写]")
	flag.BoolVar(&embedDll, "embedDll", true, "嵌入dll")
	flag.BoolVar(&embedDll, "e", true, "嵌入dll [缩写]")
	flag.BoolVar(&reduceSize, "reduceSize", false, "精减exe体积")
	flag.BoolVar(&reduceSize, "r", false, "精减exe体积 [缩写]")

	flag.Parse()

	if outName == "" {
		// 默认为当前工作目录名
		cur, _ := os.Getwd()
		outName = filepath.Base(cur)
	}
}

func genIcon() {
	icon2 := icon
	if icon == "" {
		icon2 = "icon.png"
	}
	_, err := os.Stat(icon2)
	if err != nil {
		iconFromEmbed()
	} else {
		iconFromFile(icon2)
	}
}

func iconFromFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	png2ico(f)
}

func iconFromEmbed() {
	f, err := embedFiles.ReadFile("files/applogo.ico")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("icon.ico", f, 0777)
}

func png2ico(fin io.Reader) {
	m0, err := png.Decode(fin)
	if err != nil {
		panic(err)
	}
	icoFile, _ := os.Create("icon.ico")
	defer icoFile.Close()
	ico.Encode(icoFile, m0)
}

func copyManifest() {
	f, err := embedFiles.ReadFile("files/app.manifest")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("app.manifest", f, 0777)
}

func createSyso() {
	c := coff.New()

	icon := &syso.FileResource{Name: "MAINICON", Path: "icon.ico"}
	if err := syso.EmbedIcon(c, icon); err != nil {
		panic(err)
	}

	manifest := &syso.FileResource{ID: 1, Path: "app.manifest"}
	if err := syso.EmbedManifest(c, manifest); err != nil {
		panic(err)
	}

	fout, err := os.Create("app.syso")
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	if _, err := c.WriteTo(fout); err != nil {
		panic(err)
	}
	// println("生成syso成功！")
}

func build() {
	params := make([]string, 0)
	params = append(params, "build")
	params = append(params, "-o")
	params = append(params, outName+".exe")

	cmdStr := "go build -o " + outName + ".exe"

	idflags := ""
	if !showConsole {
		idflags += "-H windowsgui"
	}
	if reduceSize {
		idflags += " -s -w"
	}
	if idflags != "" {
		params = append(params, "-ldflags")
		params = append(params, idflags)
		cmdStr += " -ldflags=\"" + idflags + "\""
	}

	if embedDll {
		params = append(params, "-tags")
		params = append(params, "tempdll")
		cmdStr += " -tags tempdll"
	}
	println("cmd:", cmdStr)

	cmd := exec.Command("go", params...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		println(err.Error(), stderr.String())
	} else {
		output := out.String()
		if output != "" {
			println(output)
		}
		println("生成 " + outName + ".exe 成功")
	}
}

// 清理app.manifest icon.ico
func clear() {
	os.Remove("app.manifest")
	os.Remove("icon.ico")
	os.Remove("app.syso")
}
