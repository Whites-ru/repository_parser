package parserepository

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	//sv "github.com/Masterminds/semver/v3"
	//sv "github.com/blang/semver/v4"
	//"github.com/hashicorp/go-version"
	sv "github.com/knqyf263/go-rpm-version"
	sl "golang.org/x/exp/slices"
)

type Packagesets struct {
	Length      int64
	Packagesets *[]string
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
	Branch_one                  string     `json:"branch_one"`
	Branch_two                  string     `json:"branch_two"`
	Arch_one                    string     `json:"arch_one"`
	Arch_two                    string     `json:"arch_two"`
	Packages_not_in_two         *[]Package `json:"packages_not_in_two"`
	Packages_not_in_one         *[]Package `json:"packages_not_in_one"`
	Packages_with_hight_version *[]Package `json:"packages_with_hight_version"`
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
	active_packagesets_url string
	all_pkgset_archs_url   string
	package_list_url       string

	pcl_1             []Package
	pcl_2             []Package
	result_arr        Result
	result_not_in_two Result_not_in_two
	result_not_in_one Result_not_in_one
	result_in_one     Result_in_one_and_two
	result_in_two     Result_in_one_and_two
	result_versions   Result_versions
	wg                sync.WaitGroup
	wg2               sync.WaitGroup
	threads_count     int

	package_found       []string
	package_found_store atomic.Value
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

func (rv *Result_versions) add_packages_with_hight_version(pck []Package) {
	rv.mu.Lock()
	defer rv.mu.Unlock()
	rv.Packages_with_hight_version = append(rv.Packages_with_hight_version, pck...)
}

func Set_api_urls(active_packagesets, all_pkgset_archs, package_list string) bool {
	res := false
	if len(strings.TrimSpace(active_packagesets)) > 0 && len(strings.TrimSpace(all_pkgset_archs)) > 0 && len(strings.TrimSpace(package_list)) > 0 {
		active_packagesets_url = strings.TrimSpace(active_packagesets)
		all_pkgset_archs_url = strings.TrimSpace(all_pkgset_archs)
		package_list_url = strings.TrimSpace(package_list)
		res = true
	}
	return res
}

func Get_package_sets() (bool, *[]string) {
	is_ok := false
	var res *[]string

	if len(strings.TrimSpace(active_packagesets_url)) > 0 {

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
	}
	return is_ok, res
}

func Get_package_set_archs(branch string) (bool, []Arch) {
	is_ok := false
	var res []Arch

	if len(strings.TrimSpace(all_pkgset_archs_url)) > 0 {

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
	}
	return is_ok, res
}

func Get_package_list(branch, arch string) (bool, []Package) {
	is_ok := false
	var res []Package

	if len(strings.TrimSpace(package_list_url)) > 0 {

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
	}
	return is_ok, res
}

func Get_package_found() []string {
	val := package_found_store.Load()
	if val == nil {
		return []string{}
	}
	return val.([]string)
}

func find_packages_vers(n_start, n_end, n2 int) {
	fmt.Println("thread n_start=" + strconv.Itoa(n_start) + " n_end=" + strconv.Itoa(n_end) + " is start at " + time.Now().Local().String())
	var pkgs []Package
	for i := n_start; i < n_end; i++ {

		// //github.com/hashicorp/go-version
		// v1, err := version.NewVersion(result_in_one.Packages[i].Version)
		// if err == nil {
		// 	for j := 0; j < n2; j++ {

		// 		if Get_package_found() == nil || !sl.Contains(Get_package_found(), result_in_one.Packages[i].Name) {
		// 			v2, err2 := version.NewVersion(result_in_two.Packages[j].Version)

		// 			if err2 == nil {

		// 				if result_in_one.Packages[i].Name == result_in_two.Packages[j].Name && v1.GreaterThan(v2) {
		// 					pck := result_in_one.Packages[i]
		// 					//result_versions.add_packages_with_hight_version(pck)

		// 					pkgs = append(pkgs, pck)

		// 					package_found_store.Store(append(package_found, pck.Name))

		// 					break
		// 				}
		// 			} else {
		// 				fmt.Print("err2: ")
		// 				fmt.Println(err2)
		// 			}
		// 		}
		// 	}
		// } else {
		// 	fmt.Print("err1: ")
		// 	fmt.Println(err)
		// }

		// //https://github.com/Masterminds/semver
		// v1, err := sv.NewVersion(result_in_one.Packages[i].Version)
		// if err == nil {
		// 	for j := 0; j < n2; j++ {

		// 		if Get_package_found() == nil || !sl.Contains(Get_package_found(), result_in_one.Packages[i].Name) {
		// 			v2, err2 := sv.NewVersion(result_in_two.Packages[j].Version)

		// 			if err2 == nil {

		// 				if result_in_one.Packages[i].Name == result_in_two.Packages[j].Name && v1.GreaterThan(v2) {
		// 					pck := result_in_one.Packages[i]
		// 					//result_versions.add_packages_with_hight_version(pck)

		// 					pkgs = append(pkgs, pck)

		// 					package_found_store.Store(append(package_found, pck.Name))

		// 					break
		// 				}
		// 			} else {
		// 				fmt.Print("err2: ")
		// 				fmt.Println(err2)
		// 			}
		// 		}
		// 	}
		// } else {
		// 	fmt.Print("err1: ")
		// 	fmt.Println(err)
		// }

		//golang.org/x/exp/slices
		// for j := 0; j < n2; j++ {

		// 	if Get_package_found() == nil || !sl.Contains(Get_package_found(), result_in_one.Packages[i].Name) {

		// 		if result_in_one.Packages[i].Name == result_in_two.Packages[j].Name && sv.Compare("v"+result_in_one.Packages[i].Version, "v"+result_in_two.Packages[i].Version) == +1 {
		// 			pck := result_in_one.Packages[i]
		// 			//result_versions.add_packages_with_hight_version(pck)

		// 			pkgs = append(pkgs, pck)

		// 			package_found_store.Store(append(package_found, pck.Name))

		// 			break
		// 		}

		// 	}
		// }

		//github.com/blang/semver/v4
		/*v1, err := sv.New(result_in_one.Packages[i].Version)
		if err == nil {
			for j := 0; j < n2; j++ {

				if Get_package_found() == nil || !sl.Contains(Get_package_found(), result_in_one.Packages[i].Name) {
					v2, err2 := sv.New(result_in_two.Packages[j].Version)

					if err2 == nil {

						if result_in_one.Packages[i].Name == result_in_two.Packages[j].Name && v1.Compare(*v2) == 1 {
							pck := result_in_one.Packages[i]
							//result_versions.add_packages_with_hight_version(pck)

							pkgs = append(pkgs, pck)

							package_found_store.Store(append(package_found, pck.Name))

							break
						}
					}
				}
			}
		}*/

		//github.com/knqyf263/go-rpm-version
		v1 := sv.NewVersion(result_in_one.Packages[i].Version)
		//fmt.Println(v1.Version())
		for j := 0; j < n2; j++ {

			if Get_package_found() == nil || !sl.Contains(Get_package_found(), result_in_one.Packages[i].Name) {
				v2 := sv.NewVersion(result_in_two.Packages[j].Version)
				//fmt.Println(v2.Version())
				if result_in_one.Packages[i].Name == result_in_two.Packages[j].Name && v1.GreaterThan(v2) {
					pck := result_in_one.Packages[i]
					//result_versions.add_packages_with_hight_version(pck)

					pkgs = append(pkgs, pck)

					package_found_store.Store(append(package_found, pck.Name))

					break
				}

			}
		}

	}
	result_versions.add_packages_with_hight_version(pkgs)

	fmt.Println("thread n_start=" + strconv.Itoa(n_start) + " n_end=" + strconv.Itoa(n_end) + " is END at " + time.Now().Local().String())
	wg2.Done()
}

func Find_packages(operation int) {
	is_find := false

	pcl_1_len := len(pcl_1)
	pcl_2_len := len(pcl_2)

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
			}
		}

		/*все пакеты, которые есть в 2-й но нет во 1-й*/
	} else if operation == 2 {
		is_find = false
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
			}
		}
		/*все пакеты, которые есть в 1-й и во 2-й*/
	} else if operation == 3 {
		for i := 0; i < pcl_1_len; i++ {
			for j := 0; j < pcl_2_len; j++ {

				if pcl_1[i].Name == pcl_2[j].Name {
					result_in_one.add_package(pcl_1[i])
					result_in_two.add_package(pcl_2[j])
					break
				}
			}
		}

		/*все пакеты, version-release которых больше в 1-й чем во 2-й*/
	} else if operation == 4 {
		var n_start, n_end, n_go int = 0, 0, 0
		pcl_1_len = len(result_in_one.Packages)
		pcl_2_len = len(result_in_two.Packages)
		n_go = int(math.Floor(float64(pcl_1_len) / float64(threads_count)))
		n_end = n_go
		wg2.Add(threads_count)
		fmt.Println(threads_count)
		for i := 0; i < threads_count; i++ {
			fmt.Println("thread no: " + strconv.Itoa(i) + " n_start=" + strconv.Itoa(n_start) + " n_end=" + strconv.Itoa(n_end))
			go find_packages_vers(n_start, n_end, pcl_2_len)

			n_start = n_start + n_go
			n_end = n_end + n_go
		}
		wg2.Wait()
	}

	if operation == 3 {
		Find_packages(4)
	} else {
		wg.Done()
	}
}

func Get_result(branch_one, branch_two, arch_one, arch_two string, thread_count int) (bool, []byte) {
	var res []byte
	var err error
	is_ok := false
	var is_ok_1, is_ok_2 bool = false, false

	if len(branch_one) > 0 && len(branch_two) > 0 && len(arch_one) > 0 && len(arch_two) > 0 {

		is_ok_1, pcl_1 = Get_package_list(branch_one, arch_one)
		is_ok_2, pcl_2 = Get_package_list(branch_two, arch_two)

		if is_ok_1 && is_ok_2 {
			fmt.Println("[] Start processing with parameters branch_one=\"" + branch_one + "\", arch_one=\"" + arch_one + "\", branch_two=\"" + branch_two + "\", arch_two=\"" + arch_two + "\"...")

			if thread_count > 0 {
				threads_count = thread_count
			} else {
				threads_count = runtime.NumCPU()
			}

			wg.Add(3)
			go Find_packages(1)
			go Find_packages(2)
			go Find_packages(3)
			wg.Wait()

			result_arr.Branch_one = branch_one
			result_arr.Branch_two = branch_two
			result_arr.Arch_one = arch_one
			result_arr.Arch_two = arch_two
			result_arr.Packages_not_in_one = &result_not_in_one.Packages_not_in_one
			result_arr.Packages_not_in_two = &result_not_in_two.Packages_not_in_two
			result_arr.Packages_with_hight_version = &result_versions.Packages_with_hight_version

			fmt.Println("len vers=" + strconv.Itoa(len(result_versions.Packages_with_hight_version)))

			res, err = json.Marshal(result_arr)
			if err == nil {
				is_ok = true
			}
		}
	}
	return is_ok, res
}
