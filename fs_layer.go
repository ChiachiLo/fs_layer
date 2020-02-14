package main

import "fmt"
import "path"
import "path/filepath"
import "os/exec"

// import "flag"
import "github.com/alexflint/go-arg"
import "strings"

// import log "github.com/sirupsen/logrus"
import "os"
import "time"
import "io/ioutil"

var (
	print_log = false
)

type MountCmd struct {
	BaseDir string   `arg:"-b" help:"Set base dir."`
	Layers  []string `arg:"-l" help:"Set level list."`
}

type UmountCmd struct {
	OutputName string `arg:"-o" help:"Set output work dir name."`
}

type DeleteCmd struct {
	Layers []string `arg:"-l" help:"Set level list."`
}

type DeleteAllCmd struct {
}

type ListCmd struct {
}

var args struct {
	Mount     *MountCmd     `arg:"subcommand:mount"`
	Umount    *UmountCmd    `arg:"subcommand:umount"`
	Delete    *DeleteCmd    `arg:"subcommand:delete"`
	DeleteAll *DeleteAllCmd `arg:"subcommand:cleanall"`
	List      *ListCmd      `arg:"subcommand:list"`
	Output    string        `arg:"required,-o" help:"Set output work dir name."`
}

func D(msg string) {
	if print_log {
		fmt.Println(fmt.Sprintf("D: %v", msg))
	}
}
func I(msg string) {
	fmt.Println(fmt.Sprintf("I: %v", msg))
}

func E(msg string) {
	fmt.Println(fmt.Sprintf("E: %v", msg))
}

func get_root_path(dest_dir string) string {
	path, _ := filepath.Abs(fmt.Sprintf(".fs_layer_%v", dest_dir))
	return path
}

func get_layer_path(dest_dir string, layer string) string {
	if layer != "" {
		return path.Join(get_root_path(dest_dir), "layer_dir", layer)
	}
	return path.Join(get_root_path(dest_dir), "layer_dir")

}

func get_work_path(dest_dir string) string {
	return path.Join(get_root_path(dest_dir), "work_dir")
}

func parser_layer_ary(layer_ary []string) []string {
	var new_plyer_ary []string
	for _, layer := range layer_ary {
		l_ary := strings.Split(layer, ":")
		new_plyer_ary = append(new_plyer_ary, l_ary...)
		D(fmt.Sprintf("new_plyer_ary:%v", new_plyer_ary))
	}
	return new_plyer_ary
}

func layers_id_to_realy_path(dest_dir string, layer_ary []string) []string {
	var ret_ary []string

	for _, layer_id := range parser_layer_ary(layer_ary) {
		dir_path := get_layer_path(dest_dir, layer_id)
		ret_ary = append(ret_ary, dir_path)
	}
	D(fmt.Sprintf("realy path:%v", ret_ary))
	return ret_ary
}

func mkdir(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err == nil {
			return true
		}
		return false
	}
	return true
}

func mkdirs(path_ary []string) bool {
	for _, path := range path_ary {
		mkdir(path)
	}
	return true
}

func mount(base_dir string, layer_ary []string, dest_dir string) bool {
	I(fmt.Sprintf("Mount, base:%v, layer:%v, dest:%v",
		base_dir, layer_ary, dest_dir))

	layer_ary = layers_id_to_realy_path(dest_dir, layer_ary)
	mkdirs(layer_ary)
	mkdir(get_work_path(dest_dir))
	mkdir(dest_dir)

	// # sudo mount -t overlay -o lowerdir=${src_dir},upperdir=${upper_dir},workdir=${work_dir} none ${dest_dir}
	var lowerdir_ary []string
	base_dir, _ = filepath.Abs(base_dir)
	lowerdir_ary = append(lowerdir_ary,
		base_dir)
	lowerdir_ary = append(
		lowerdir_ary,
		layer_ary[:len(layer_ary)-1]...,
	)

	D(fmt.Sprintf("lowerdir_ary:%v", lowerdir_ary))
	lowerdir := strings.Join(lowerdir_ary, ":")
	upperdir := layer_ary[len(layer_ary)-1]
	workdir := get_work_path(dest_dir)
	// workdir =  get_work_path(dest_dir)
	cmd_ary := []string{"sudo", "mount", "-t", "overlay", "-o"}

	// cmd_ary = [ 'sudo', 'mount', '-t', 'overlay', '-o']
	cmd_ary = append(cmd_ary, fmt.Sprintf("lowerdir=%v,upperdir=%v,workdir=%v", lowerdir, upperdir, workdir))

	// cmd_ary.append('lowerdir={},upperdir={},workdir={}'.format(lowerdir, upperdir, workdir ))
	dest_dir, _ = filepath.Abs(dest_dir)
	cmd_ary = append(cmd_ary, "none", dest_dir)
	// cmd_ary.extend(['none', os.path.abspath(dest_dir)])
	D(fmt.Sprintf("cmd:%v", cmd_ary))
	_, err := exec.Command(cmd_ary[0], cmd_ary[1:]...).Output()
	if err != nil {
		D(fmt.Sprintf("Run mount command fail..."))
		D(fmt.Sprintf("%v", err))
		return false
	}
	return true
	// run(cmd_ary)
}

func umount(folder string) bool {
	I(fmt.Sprintf("Umount, dest:%v",
		folder))

	cmd_ary := []string{"sudo", "umount", "-l", folder}

	_, err := exec.Command(cmd_ary[0], cmd_ary[1:]...).Output()
	if err == nil {
		return true
	}

	time.Sleep(1 * time.Second)
	cmd_ary = []string{"sudo", "umount", "-f", folder}
	_, err = exec.Command(cmd_ary[0], cmd_ary[1:]...).Output()
	if err != nil {
		return false
	}
	return true
}

func clean(dest_dir string) bool {
	I(fmt.Sprintf("Clean, dest:%v", dest_dir))
	cmd_ary := []string{"sudo", "rm", "-rf", get_root_path(dest_dir)}
	exec.Command(cmd_ary[0], cmd_ary[1:]...).Output()

	cmd_ary = []string{"sudo", "rm", "-rf", dest_dir}
	exec.Command(cmd_ary[0], cmd_ary[1:]...).Output()
	return true
}

func list_layer(dest_dir string) {
	I(fmt.Sprintf("List layer, dest:%v", dest_dir))

	dest_dir = get_layer_path(dest_dir, "")
	files, err := ioutil.ReadDir(dest_dir)
	if err != nil {
		E(fmt.Sprintf("List layer fail:%v", err))
		return
	}
	for i, f := range files {
		fmt.Println(fmt.Sprintf("%v:%v", i, f.Name()))
	}
}

func test_parser_layer_ary() {
	testv := []string{"test1", "test2:test3"}
	parser_layer_ary(testv)
}

func test_mount() {
	tags := []string{"tag1:tag2", "tag3"}
	mount("test_base", tags, "work_dir")
}

func test() {
	test_mount()
}

func main() {
	// test()
	arg.MustParse(&args)

	switch {
	case args.Mount != nil:
		ret := mount(args.Mount.BaseDir, args.Mount.Layers, args.Output)
		if ret == false {
			os.Exit(1)
		}
	case args.Umount != nil:
		ret := umount(args.Output)
		if ret == false {
			os.Exit(2)
		}
	case args.DeleteAll != nil:
		clean(args.Output)
	case args.List != nil:
		list_layer(args.Output)
	}
	os.Exit(0)

}
