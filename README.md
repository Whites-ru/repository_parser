# Repository parser
## Repository parsing utility with branch comparison.

### Description
Comparison conditions:
1) All packages that are in the 1st branch, but not in the 2nd branch
2) All packages that are in the 2nd branch, but they are not in the 1st branch
3) All packages whose version-release is larger in the 1st branch than in the 2nd branch.

The summary data contains:
1) The name of the first branch is *Branch_one*
2) The name of the second branch is *Branch_two*
3) The name of the architecture of the first branch is *Arch_one*
4) The name of the architecture of the second branch is *Arch_two*
5) The list of packages falling under the first condition is *Packages_not_in_two*
6) The package list falling under the second condition is *Packages_not_in_one*
7) The list of packages falling under the third condition is *Packages_with_hight_version*

### Usage
The program can be started in two modes:
1) View Help
2) Repository parsing

Starting in help mode is carried out with the start of the program with the *help* flag.
Starting in the working mode of parsing is carried out with the indication of flags:
- *branch_one* - the name of the first branch (mandatory flag)
- *branch_two* - the name of the second branch (mandatory flag)
- *arch_one* - the name of the first architecture (mandatory flag)
- *arch_two* - the name of the second architecture (mandatory flag)
- *output_file* - the path and name of the file in which the result of the work will be written. If the flag is not specified, the result is displayed in the console.
- *thread_count* - the number of threads used for parsing. If the flag is not set, the number of threads will be used equal to the number of processors (virtual) in the system.

### Additional info
The program is written in the Go language using a standard library and *github.com/knqyf263/go-rpm-version* library. 
Appplication can be assembled for all architectures and operating systems supported by the Go language.

Examples for building a program:
1) for Windows operating systems: *go build -ldflags = "-w -s" -o parse_repository.exe*
2) for Linux kernel operating systems: *go build -ldflags = "-w -s" -o parse_repository*