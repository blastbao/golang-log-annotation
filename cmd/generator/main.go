package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
	"github.com/blastbao/golang-log-annotation"
)

var replace = false

func init() {
	flag.BoolVar(&replace, "replace", false, "使用生成的代码替换原本的代码")
}

func main() {
	flag.Parse()
	// 遍历需要应用注解的文件夹列表
	for _, dirPath := range flag.Args() {
		// 遍历每个文件夹下的所有文件/文件夹，并对每个元素执行 overwrite 函数
		if err := filepath.Walk(dirPath, walkFn); err != nil {
			panic(err)
		}
	}
}

func Generate(dir string) error {
	if err := filepath.Walk(dir, walkFn); err != nil {
		return err
	}
	return nil
}


type Modifier struct {}

func (m *Modifier) Generate(dir string) error {
	return filepath.Walk(dir, m.walkDirFn)
}

func (m *Modifier) walkDirFn(filePath string, fileInfo os.FileInfo, err error) error {

	if fileInfo.IsDir() {
		return nil
	}

	if !strings.HasSuffix(fileInfo.Name(), ".go") {
		return nil
	}

	fileSet, fileAst, err := m.parseFileSet(filePath)
	if err != nil {
		return err
	}



}

func (m *Modifier) parseFileSet(filePath string) (*token.FileSet, *ast.File, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filePath, content, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}
	return fileSet, file, nil
}


func (m *Modifier) procLogInject(filePath string, fileSet *token.FileSet, file *ast.File) (bool, error ) {

	// 初始化处理本次文件所需的信息对象
	info := &Info{
		Filepath: filepath,
		NamedImportAdder: func(name string, path string) bool {
			return astutil.AddNamedImport(fileSet, file, name, path)
		},
	}

	astutil.AddNamedImport(fileSet, file, "xxx", filePath)

	// 遍历当前文件 ast 上的所有节点
	astutil.Apply(file, nil, func(cursor *astutil.Cursor) bool {

		node := cursor.Node()

		funcDecl, ok := node.(*ast.FuncDecl) // log 注解只用于函数
		if !ok {
			return false
		}

		if funcDecl.Doc == nil { // 如果没有注释，则直接处理下一个
			return true
		}

		// 处理 log 注解
		info.Node = cursor.Node()

		nodeModified, err := Handler.Handle(info)
		if err != nil {
			panic(err)
		}

		if nodeModified {
			modified = nodeModified
		}

		return true
	})

	return
}

func (m *Modifier) logInject(node ast.Node) (modified bool, err error) {

	// log 注解只用于函数
	funcDecl, ok := node.(*ast.FuncDecl)
	if !ok {
		return
	}

	// 如果没有注释，则直接处理下一个
	if funcDecl.Doc == nil {
		return
	}

	// 如果不是可以处理的注解，则直接返回
	doc := strings.Trim(funcDecl.Doc.Text(), "\t \n")
	if doc != "@Log()" {
		return
	}

	// 标记为已修改
	modified = true
	// 删除注释注解
	funcDecl.Doc.List = nil
	// 导入项目中 Logger 所在的包
	info.NamedImportAdder("", "logannotation/testdata/log")

}



// walkFn 函数会对每个 .go 文件处理，并调用注解处理器
func walkFn(path1 string, info os.FileInfo, err error) error {

	// 如果是文件夹，或者不是 .go 文件，则直接返回不处理
	if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
		return nil
	}

	fileSet, file, err := parseFile(path1)

	// 如果注解修改了内容，则需要生成新的代码
	if logannotation.Overwrite(path1, fileSet, file) {

		buf := &bytes.Buffer{}
		if err := format.Node(buf, fileSet, file); err != nil {
			panic(err)
		}

		// 如果不需要替换，则生成到另一个文件
		if !replace {

			// path.Dir(path1) + "/_gen/"


			lastSlashIndex := strings.LastIndex(path1, "/")
			genDirPath := path1[:lastSlashIndex] + "/_gen/"
			if err := os.Mkdir(genDirPath, 0755); err != nil && os.IsNotExist(err) {
				panic(err)
			}
			path1 = genDirPath + path1[lastSlashIndex+1:]
		}

		if err := ioutil.WriteFile(path1, buf.Bytes(), info.Mode()); err != nil {
			panic(err)
		}
	}

	return nil
}

func parseFile(filepath string) (*token.FileSet, *ast.File, error) {

	source, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, nil, err
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filepath, source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}

	return fileSet, file, nil
}
