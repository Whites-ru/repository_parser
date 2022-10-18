package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Packagesets struct {
	Length      int64
	Packagesets []*string
}

type Arch struct {
	Arch  string
	Count int64
}

type Archs struct {
	Length int64
	Archs  []Arch
}

type Args struct {
	Arch string
}

type Package struct {
	Name      string
	Epoch     int64
	Version   string
	Release   string
	Arch      string
	Disttag   string
	Buildtime int64
	Source    string
}

type Response struct {
	Request_args Args
	Length       int64
	Packages     []Package
}

type Result struct {
	Branch_one                  string
	Branch_two                  string
	Arch_one                    string
	Arch_two                    string
	Packages_not_in_two         *[]Package
	Packages_not_in_one         *[]Package
	Packages_with_hight_version *[]Package
}

type Result_versions struct {
	mu                          sync.Mutex
	Packages_with_hight_version []Package
}

type Result_not_in_two struct {
	Packages_not_in_two []Package
}

type Result_not_in_one struct {
	Packages_not_in_one []Package
}
type Result_in_one_and_two struct {
	Packages []Package
}

var (
	active_packagesets_url string = "https://rdb.altlinux.org/api/packageset/active_packagesets"
	all_pkgset_archs_url   string = "https://rdb.altlinux.org/api/site/all_pkgset_archs"
	package_list_url       string = "https://rdb.altlinux.org/api/export/branch_binary_packages"

	pcl_1 []Package
	pcl_2 []Package
	// pcl_3                 []Package
	result_arr            Result
	result_not_in_two     Result_not_in_two
	result_not_in_one     Result_not_in_one
	result_in_one_and_two Result_in_one_and_two
	result_versions       Result_versions
	wg                    sync.WaitGroup
	wg2                   sync.WaitGroup
	threads_count         int
)

func (r *Result_not_in_two) add_packages_not_in_two(pck Package) {
	r.Packages_not_in_two = append(r.Packages_not_in_two, pck)
}

func (r *Result_not_in_one) add_packages_not_in_one(pck Package) {
	r.Packages_not_in_one = append(r.Packages_not_in_one, pck)
}

func (r *Result_in_one_and_two) add_package(pck Package) {
	r.Packages = append(r.Packages, pck)
}

func (rv *Result_versions) add_packages_with_hight_version(pck Package) {
	rv.mu.Lock()
	defer rv.mu.Unlock()
	rv.Packages_with_hight_version = append(rv.Packages_with_hight_version, pck)
}

func Get_package_sets() (bool, []*string) {
	is_ok := false
	var res []*string

	resp, err := http.Get(active_packagesets_url)
	if err != nil {
		fmt.Println(err)

	} else {
		body := json.NewDecoder(resp.Body)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fmt.Printf("Response failed with status code: %d\n", resp.StatusCode)
		} else {
			if err != nil {
				fmt.Println(err)
			} else {
				var data Packagesets
				err := body.Decode(&data)
				if err != nil {
					fmt.Println(err)
				} else {
					res = data.Packagesets
					is_ok = true
				}
			}
		}
	}
	return is_ok, res
}

func Get_package_set_archs(branch string) (bool, []Arch) {
	is_ok := false
	var res []Arch

	resp, err := http.Get(all_pkgset_archs_url + "?branch=" + branch)
	if err != nil {
		fmt.Println(err)

	} else {
		body := json.NewDecoder(resp.Body)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fmt.Printf("Response failed with status code: %d\n", resp.StatusCode)
		} else {
			if err != nil {
				fmt.Println(err)
			} else {
				var data Archs
				err := body.Decode(&data)
				if err != nil {
					fmt.Println(err)
				} else {
					res = data.Archs
					is_ok = true
				}
			}
		}
	}
	return is_ok, res
}

func Get_package_list(branch, arch string) (bool, []Package) {
	is_ok := false
	var res []Package

	resp, err := http.Get(package_list_url + "/" + branch + "?arch=" + arch)
	if err != nil {
		fmt.Println(err)

	} else {
		body := json.NewDecoder(resp.Body)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fmt.Printf("Response failed with status code: %d\n", resp.StatusCode)
		} else {
			if err != nil {
				fmt.Println(err)
			} else {
				var data Response
				err := body.Decode(&data)
				if err != nil {
					fmt.Println(err)
				} else {
					res = data.Packages
					is_ok = true
				}
			}
		}
	}
	return is_ok, res
}

func find_packages_vers(n_start, n_end, n2 int /*, p1, p2 *[]Package*/) {
	var ver_maj_1, ver_min_1, ver_rel_1, ver_maj_2, ver_min_2, ver_rel_2 int64 = 0, 0, 0, 0, 0, 0

	var ver_arr []string
	var err, err2 error
	//var n int = 0

	fmt.Println("exec n_start=" + strconv.Itoa(n_start) + " n_end=" + strconv.Itoa(n_end) + "....")
	for i := n_start; i < n_end; i++ {
		ver_arr = strings.Split(result_in_one_and_two.Packages[i].Version, ".")
		//fmt.Println(ver_arr)

		ver_maj_1, err = strconv.ParseInt(ver_arr[0], 10, 64)
		if err == nil {
			if len(ver_arr) > 1 {
				ver_min_1, err = strconv.ParseInt(ver_arr[1], 10, 64)
				if err == nil {
					if len(ver_arr) >= 3 {
						ver_rel_1, err = strconv.ParseInt(ver_arr[2], 10, 64)
					}
				}
			}
		}
		if err == nil {

			for j := 0; j < n2; j++ {

				ver_arr = strings.Split(pcl_2[j].Version, ".")
				//fmt.Println(ver_arr)

				ver_maj_2, err2 = strconv.ParseInt(ver_arr[0], 10, 64)
				if err2 == nil {
					if len(ver_arr) > 1 {
						ver_min_2, err2 = strconv.ParseInt(ver_arr[1], 10, 64)
						if err2 == nil {
							if len(ver_arr) >= 3 {
								ver_rel_2, err2 = strconv.ParseInt(ver_arr[2], 10, 64)
							}
						}
					}
				}

				if err2 == nil {

					if result_in_one_and_two.Packages[i].Name == pcl_2[j].Name && ver_maj_1 > ver_maj_2 &&
						((ver_min_1 > 0 && ver_min_2 > 0 && ver_min_1 > ver_min_2) || (ver_min_1 == 0 && ver_min_2 == 0)) &&
						((ver_rel_1 > 0 && ver_rel_2 > 0 && ver_rel_1 > ver_rel_2) || (ver_rel_1 == 0 && ver_rel_2 == 0)) {
						fmt.Printf("Name 1 highter: %s ver1=%s ver2=%s \n", result_in_one_and_two.Packages[i].Name, result_in_one_and_two.Packages[i].Version, pcl_2[j].Version)
						pck := result_in_one_and_two.Packages[i]
						result_versions.add_packages_with_hight_version(pck)
						//n++
						break
					}
				}
			}
		}
	}
	//if n > 0 {
	wg2.Done()
	fmt.Println("exec ok n_start=" + strconv.Itoa(n_start) + " n_end=" + strconv.Itoa(n_end))
	//}
}

func Find_packages(operation int) bool {
	is_ok := false

	is_find := false
	n := 0

	pcl_1_len := len(pcl_1)
	pcl_2_len := len(pcl_2)

	fmt.Println("<-- operation: " + strconv.Itoa(operation) + " started at: " + time.Now().Local().String())

	/*все пакеты, которые есть в 1-й но нет во 2-й*/
	if operation == 1 {
		for i := 0; i < pcl_1_len; i++ {
			is_find = false
			for j := 0; j < pcl_2_len; j++ {

				if pcl_1[i].Name == pcl_2[j].Name {
					is_find = true
					break
				}
			}
			if !is_find {
				result_not_in_two.add_packages_not_in_two(pcl_1[i])
				n++
			}
		}

		/*все пакеты, которые есть в 2-й но нет во 1-й*/
	} else if operation == 2 {
		is_find = false
		n = 0
		for i := 0; i < pcl_2_len; i++ {
			is_find = false
			for j := 0; j < pcl_1_len; j++ {

				if pcl_2[i].Name == pcl_1[j].Name {
					is_find = true
					break
				}
			}
			if !is_find {
				result_not_in_one.add_packages_not_in_one(pcl_2[i])
				n++
			}
		}
		/*все пакеты, которые есть в 1-й и во 2-й*/
	} else if operation == 3 {
		//is_find = false
		n = 0
		for i := 0; i < pcl_1_len; i++ {
			//is_find = false
			for j := 0; j < pcl_2_len; j++ {

				if pcl_1[i].Name == pcl_2[j].Name {
					//is_find = true
					result_in_one_and_two.add_package(pcl_1[i])
					n++
					break
				}
			}
			// if is_find {
			// 	//result_in_one_and_two.add_package(pcl_1[i])
			// 	n++
			// }
		}

	} else if operation == 4 {
		//var ver_maj_1, ver_min_1, ver_rel_1, ver_maj_2, ver_min_2, ver_rel_2 int64 = 0, 0, 0, 0, 0, 0

		//var ver_arr []string
		//var err, err2 error

		var n_start, n_end, n_go int = 0, 0, 0

		pcl_1_len = len(result_in_one_and_two.Packages)

		fmt.Println(pcl_1_len)
		fmt.Println(math.Floor(float64(pcl_1_len) / float64(threads_count)))

		n_go = int(math.Floor(float64(pcl_1_len) / float64(threads_count)))

		n_end = n_go
		wg2.Add(threads_count)
		for i := 0; i < threads_count; i++ {

			go find_packages_vers(n_start, n_end, pcl_2_len)
			n_start = n_start + n_go
			n_end = n_end + n_go
		}
		wg2.Wait()
		//wg.Done()
		n++

		/*for i := 0; i < pcl_1_len; i++ {
			ver_arr = strings.Split(pcl_1[i].Version, ".")

			ver_maj_1, err = strconv.ParseInt(ver_arr[0], 10, 64)
			if err == nil {
				if len(ver_arr) > 1 {
					ver_min_1, err = strconv.ParseInt(ver_arr[1], 10, 64)
					if err == nil {
						if len(ver_arr) >= 3 {
							ver_rel_1, err = strconv.ParseInt(ver_arr[2], 10, 64)
						}
					}
				}
			}
			if err == nil {

				for j := 0; j < pcl_2_len; j++ {

					ver_arr = strings.Split(pcl_2[j].Version, ".")

					ver_maj_2, err2 = strconv.ParseInt(ver_arr[0], 10, 64)
					if err2 == nil {
						if len(ver_arr) > 1 {
							ver_min_2, err = strconv.ParseInt(ver_arr[1], 10, 64)
							if err == nil {
								if len(ver_arr) >= 3 {
									ver_rel_2, err2 = strconv.ParseInt(ver_arr[2], 10, 64)
								}
							}
						}
					}

					if err2 == nil {

						if pcl_1[i].Name == pcl_2[j].Name && ver_maj_1 > ver_maj_2 &&
							((ver_min_1 > 0 && ver_min_2 > 0 && ver_min_1 > ver_min_2) || (ver_min_1 == 0 && ver_min_2 == 0)) &&
							((ver_rel_1 > 0 && ver_rel_2 > 0 && ver_rel_1 > ver_rel_2) || (ver_rel_1 == 0 && ver_rel_2 == 0)) {
							//fmt.Printf("Name 1 highter: %s ver1=%s ver2=%s \n", pcl_1[i].Name, pcl_1[i].Version, pcl_2[j].Version)
							pck := pcl_1[i]
							result_versions.add_packages_with_hight_version(pck)
							n++
							break
						}
					}
				}
			}
		}*/
	}
	if n > 0 {
		is_ok = true

		if operation == 3 {
			Find_packages(4)
		} else {
			wg.Done()
		}

		fmt.Println("<-- operation " + strconv.Itoa(operation) + " complete at: " + time.Now().Local().String())
	}
	return is_ok
}

func main() {
	is_ok, pcs := Get_package_sets()
	if is_ok {
		for i := 0; i < len(pcs); i++ {
			fmt.Printf("branch name: %s \n", *pcs[i])
		}
	}

	var archs []Arch
	is_ok, archs = Get_package_set_archs("p10")
	if is_ok {
		for i := 0; i < len(archs); i++ {
			fmt.Printf("arch: %s \n", archs[i].Arch)
		}
	}

	var is_ok_1, is_ok_2 bool = false, false

	branch_one := "p9"
	branch_two := "p10"
	arch_one := "x86_64"
	arch_two := "x86_64"

	is_ok_1, pcl_1 = Get_package_list(branch_one, arch_one)
	is_ok_2, pcl_2 = Get_package_list(branch_two, arch_two)

	if is_ok_1 && is_ok_2 {

		//var result_arr Result
		//pcl_1_len := len(pcl_1)
		//pcl_2_len := len(pcl_2)

		threads_count = runtime.NumCPU()

		wg.Add(3)
		go Find_packages(1)
		go Find_packages(2)
		go Find_packages(3)
		wg.Wait()

		//fmt.Println(pcl_1_len)
		//fmt.Println(pcl_2_len)

		result_arr.Branch_one = branch_one
		result_arr.Branch_two = branch_two
		result_arr.Arch_one = arch_one
		result_arr.Arch_two = arch_two
		result_arr.Packages_not_in_one = &result_not_in_one.Packages_not_in_one
		result_arr.Packages_not_in_two = &result_not_in_two.Packages_not_in_two
		result_arr.Packages_with_hight_version = &result_versions.Packages_with_hight_version

		res_json, _ := json.Marshal(result_arr)

		out_file, err_file := os.OpenFile("result.json", os.O_CREATE|os.O_WRONLY, 0660)
		if err_file != nil {
			fmt.Println(err_file)
		} else {
			defer out_file.Close()
			out_file.Write(res_json)
		}

	}

}
