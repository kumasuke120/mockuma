package internal

import "fmt"

const (
	AppName       = "MocKuma"
	VersionNumber = "1.1.3"
	author        = "kumasuke120<bearcomingx@gmail.com>"
	github        = "https://github.com/kumasuke120/mockuma"
	gitee         = "https://gitee.com/kumasuke/mockuma"
)

func PrintVersion() {
	fmt.Println(` _______              __  __                       `)
	fmt.Println(`|   |   |.-----.----.|  |/  |.--.--.--------.---.-.`)
	fmt.Println(`|       ||  _  |  __||     < |  |  |        |  _  |`)
	fmt.Println(`|__|_|__||_____|____||__|\__||_____|__|__|__|___._|`)
	fmt.Println()
	fmt.Printf("Version\t: %s\n", VersionNumber)
	fmt.Printf("Author\t: %s\n", author)
	fmt.Printf("GitHub\t: %s\n", github)
	fmt.Printf("Gitee\t: %s\n", gitee)
}
