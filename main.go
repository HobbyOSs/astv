package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/HobbyOSs/astv/internal/gen_handler"
	"github.com/HobbyOSs/astv/internal/option"
)

const version = "0.1.0"

func main() {
	opts := &option.Options{}

	flag.StringVar(&opts.Dir, "d", ".", "読み取り対象ディレクトリを指定する、デフォルトはカレントディレクトリ")
	flag.StringVar(&opts.Asts, "ast", "", "カンマ区切りのASTの型一覧")
	flag.BoolVar(&opts.FormatCode, "format", false, "コード生成後にコードをフォーマットする")
	flag.BoolVar(&opts.ShowVersion, "v", false, "バージョン情報を表示する")
	flag.Parse()

	// バージョン情報を表示するオプションが指定された場合
	if opts.ShowVersion {
		fmt.Println("Version:", version)
		os.Exit(0)
	}
	if len(opts.Asts) > 0 {
		opts.AstTypes = strings.Split(opts.Asts, ",")
	}

	if err := gen_handler.GenHandler(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
