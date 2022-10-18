package main

import (
	"flag"
	"fmt"
	"os"
	pr "parse_repository"
	"runtime"
	"strings"
)

func get_help(by_error bool) {
	if by_error {
		fmt.Println("Invalid call parameters.")
	}
	fmt.Println("Available branch names:")
	is_ok, pcs := pr.Get_package_sets()
	if is_ok {
		fmt.Println(strings.Join(*pcs, " "))
	}

	fmt.Println("\nAvailable archs:")
	var archs []pr.Arch
	is_ok, archs = pr.Get_package_set_archs("p10")
	if is_ok {
		for i := 0; i < len(archs); i++ {
			fmt.Printf("%s ", archs[i].Arch)
		}
	}
	fmt.Println("\n\nTo use set flags \"branch_one\", \"branch_two\" from available branch names and set flags \"arch_one\", \"arch_two\" from available archs; if you set flag \"output_file\" result save to file or view result in console; set flag \"thread_count\" to set count of thread to use. Or use \"help\" parameter for show help.")
	fmt.Println("Example usage: \"-branch_one=p10 -branch_two=p9 -arch_one=x86_64 -arch_two=x86_64 -output_file=result.json -thread_count=2\"")
}

func main() {
	if len(os.Args) < 1 {
		get_help(true)
		os.Exit(2)
	}

	if pr.Set_api_urls("https://rdb.altlinux.org/api/packageset/active_packagesets",
		"https://rdb.altlinux.org/api/site/all_pkgset_archs",
		"https://rdb.altlinux.org/api/export/branch_binary_packages") {

		if os.Args[1] == "help" {
			get_help(false)
		} else {

			branch_one := flag.String("branch_one", "", "a string to set one branch name")
			branch_two := flag.String("branch_two", "", "a string to set two branch name")
			arch_one := flag.String("arch_one", "", "a string to set one arch name")
			arch_two := flag.String("arch_two", "", "a string to set two arch name")

			output_file := flag.String("output_file", "", "a string to set path of output file")
			thread_count := flag.Int("thread_count", runtime.NumCPU(), "a number of thread to use")

			flag.Parse()

			ok, res := pr.Get_result(strings.TrimSpace(*branch_one), strings.TrimSpace(*branch_two), strings.TrimSpace(*arch_one), strings.TrimSpace(*arch_two), *thread_count)
			if ok {
				if len(strings.TrimSpace(*output_file)) > 0 {
					fmt.Println("[] Processing complete. Save result to file..")
					out_file, err_file := os.OpenFile(strings.TrimSpace(*output_file), os.O_CREATE|os.O_WRONLY, 0660)
					if err_file != nil {
						fmt.Println(err_file)
					} else {
						defer out_file.Close()
						out_file.Write(res)
						fmt.Println("[] Complete saving result to file.")
					}
				} else {
					fmt.Println("[] Processing complete.")
					fmt.Println(string(res))
				}
			} else {
				get_help(true)
			}
		}

	}
}
