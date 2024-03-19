package main

import (
	"fmt"
	"flag"
	"os"
	"github.com/HobbyOSs/astv/internal/gen_handler"
)

const version = "0.1.0"

func main() {
	var showVersion bool
	var dir string

	flag.StringVar(&dir, "d", ".", "読み取り対象ディレクトリを指定する、デフォルトはカレントディレクトリ")
	flag.BoolVar(&showVersion, "v", false, "バージョン情報を表示する")
	flag.Parse()

	// バージョン情報を表示するオプションが指定された場合
	if showVersion {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	if err := gen_handler.GenHandler(dir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
