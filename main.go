// BackupFolder project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"linbo.ga/toolfunc"
)

var needcfsm = make(map[string]bool, 0)
var needcdsm = make(map[string]bool, 0)
var hostdirsm = make(map[string]bool, 0)
var cpcount int
var totalcount int
var targetmodgreatthansrc []string

func GetNeedCleanByRulesFiles(rulepath string, files, dirs, rootdirs []string) (needdelfiles, needdeldirs []string) {
	if toolfunc.FileExists(rulepath) == false {
		rulepath = `../CleanDirectoryByRules/CleanDirectoryByRules.rules`
	}
	if toolfunc.FileExists(rulepath) == false {
		exename := toolfunc.GetExeName()
		rulepath = toolfunc.AppDir() + exename + ".rules"
	}
	if toolfunc.FileExists(rulepath) == false {
		ioutil.WriteFile(rulepath, []byte(`#-filer:regex for remove file
#-dirr:regex for remove directiory
#-direxr:regex`+"`"+`exclude_replace_regex    for remove directiory and with exclude path regex
#-direqr:regex`+"`"+`regex_replace=match=regex_replace    if match equation regexp are eqaul delete first regex match directory
#-filepath:path for remove file
#-dirpath:path for remove directiory
#-dircfr:delete directory with child file regex
#-pdircfr:n:delete dir's parnt with child file regex,n is layer of the parent directory,0 or 1 is current file parent directory
#-dircfs:delete directory by child file names,names seporate by '|'
#-dircfds:delete directory by child file and folder names,names seporate by '|'
#-pdircfs:delete directory's parent by child file names,names seporate by '|'
#-pdircfds:n:delete directory's parent by child file and folder names,names seporate by '|'
#-exfile: exclude delete file path
#-exdir: exclude delete directory path
`), 0666)
	}
	log.Println("rulepath", rulepath)
	log.Println("files count", len(files))
	log.Println("dirs count", len(dirs))
	ctt, _ := ioutil.ReadFile(rulepath)
	lis := strings.Split(string(ctt), "\n")
	ruexfilesm := make(map[string]bool, 0)
	ruexdirsm := make(map[string]bool, 0)
	for i := 0; i < len(lis); i++ {
		li := strings.Trim(lis[i], " \r\n\t")
		if len(li) > 0 {
			if li[0] == '#' {
				continue
			}
			if strings.HasPrefix(li, "-exfile:") {
				exfi := strings.Trim(li[len("-exfile:"):], " \r\n\t")
				if exfi != "" {
					ruexfilesm[rootdirs[0]+toolfunc.StdFilePath(exfi)] = true
				}
			} else if strings.HasPrefix(li, "-exdir:") {
				exfi := strings.Trim(li[len("-exdir:"):], " \r\n\t")
				if exfi != "" {
					ruexdirsm[rootdirs[0]+toolfunc.StdDir(exfi)] = true
				}
			}
		}
	}
	for i := 0; i < len(lis); i++ {
		li := strings.Trim(lis[i], " \r\n\t")
		if len(li) > 0 {
			if li[0] == '#' {
				continue
			}
			if strings.HasPrefix(li, "-filer:") {
				re, ree := regexp.Compile(li[len("-filer:"):])
				if ree == nil {
					for _, path := range files {
						ma := re.FindAllStringSubmatch(path, 1)
						if len(ma) > 0 {
							if ruexfilesm[path] == false {
								//os.Remove(path)
								for _, rdir := range rootdirs {
									if strings.HasPrefix(path, rdir) {
										needdelfiles = append(needdelfiles, path[len(rdir):])
										break
									}
								}
							}

						}
					}
				}
			} else if strings.HasPrefix(li, "-dircfr:") {
				re, ree := regexp.Compile(li[len("-dircfr:"):])
				if ree == nil {
					for _, path := range files {
						filename := path[strings.LastIndexAny(path, "/\\")+1:]
						ma := re.FindAllStringSubmatch(filename, 1)
						if len(ma) > 0 {
							deldir := path[:strings.LastIndexAny(path, "/\\")+1]
							if hostdirsm[deldir] == false {
								if ruexdirsm[deldir] == false {
									//os.RemoveAll(deldir)
									for _, rdir := range rootdirs {
										if strings.HasPrefix(path, rdir) {
											needdeldirs = append(needdeldirs, deldir[len(rdir):])
											break
										}
									}
								}

							}
						}
					}
				}
			} else if strings.HasPrefix(li, "-pdircfr:") {
				rusub := li[len("-pdircfr:"):]
				parentlayerid := toolfunc.Atoi(rusub)
				rusub = rusub[strings.Index(rusub, ":")+1:]
				re, ree := regexp.Compile(rusub)
				if ree == nil {
					for _, path := range files {
						filename := path[strings.LastIndexAny(path, "/\\")+1:]
						ma := re.FindAllStringSubmatch(filename, 1)
						if len(ma) > 0 {
							deldir := path[:strings.LastIndexAny(path, "/\\")+1]
							if hostdirsm[deldir] == false {
								bok := true
								for pidi := 1; pidi < parentlayerid; pidi++ {
									deldir = strings.TrimRight(deldir, "/\\")
									if strings.LastIndexAny(deldir, "/\\") != -1 {
										deldir = deldir[:strings.LastIndexAny(deldir, "/\\")+1]
									} else {
										bok = false
										break
									}
								}
								if bok {
									for dir4, _ := range hostdirsm {
										if strings.HasPrefix(deldir, dir4) {
											if ruexdirsm[deldir] == false {
												//os.RemoveAll(deldir)
												for _, rdir := range rootdirs {
													if strings.HasPrefix(path, rdir) {
														needdeldirs = append(needdeldirs, deldir[len(rdir):])
														break
													}
												}
											}

											break
										}
									}
								}
							}
						}
					}
				}
			} else if strings.HasPrefix(li, "-dircfs:") {
				rusub := li[len("-dircfs:"):]
				childfilenames := strings.Split(rusub, "|")
				childfilenamesm := make(map[string]bool, 0)
				for _, name := range childfilenames {
					name = strings.Trim(name, " \r\n\t")
					if name != "" {
						childfilenamesm[name] = true
					}
				}
				for _, path := range files {
					filename := path[strings.LastIndexAny(path, "/\\")+1:]
					if childfilenamesm[filename] {
						bdel := true
						deldir := path[:strings.LastIndexAny(path, "/\\")+1]
						for name, _ := range childfilenamesm {
							if toolfunc.FileExists(deldir+name) == false {
								bdel = false
								break
							}
						}
						if bdel {
							if ruexdirsm[deldir] == false {
								//os.RemoveAll(deldir)
								for _, rdir := range rootdirs {
									if strings.HasPrefix(path, rdir) {
										needdeldirs = append(needdeldirs, deldir[len(rdir):])
										break
									}
								}
							}

						}
					}
				}
			} else if strings.HasPrefix(li, "-dircfds:") {
				rusub := li[len("-dircfds:"):]
				childfilenames := strings.Split(rusub, "|")
				childfilenamesm := make(map[string]bool, 0)
				for _, name := range childfilenames {
					name = strings.Trim(name, " \r\n\t")
					if name != "" {
						childfilenamesm[name] = true
					}
				}
				for _, path := range files {
					filename := path[strings.LastIndexAny(path, "/\\")+1:]
					if childfilenamesm[filename] {
						bdel := true
						deldir := path[:strings.LastIndexAny(path, "/\\")+1]
						for name, _ := range childfilenamesm {
							if !(toolfunc.FileExists(deldir+name) || toolfunc.FolderExists(deldir+name)) {
								bdel = false
								break
							}
						}
						if bdel {
							if ruexdirsm[deldir] == false {
								//os.RemoveAll(deldir)
								for _, rdir := range rootdirs {
									if strings.HasPrefix(path, rdir) {
										needdeldirs = append(needdeldirs, deldir[len(rdir):])
										break
									}
								}
							}

						}
					}
				}
			} else if strings.HasPrefix(li, "-pdircfs:") {
				rusub := li[len("-pdircfs:"):]
				parentlayerid := toolfunc.Atoi(rusub)
				rusub = rusub[strings.Index(rusub, ":")+1:]
				childfilenames := strings.Split(rusub, "|")
				childfilenamesm := make(map[string]bool, 0)
				for _, name := range childfilenames {
					name = strings.Trim(name, " \r\n\t")
					if name != "" {
						childfilenamesm[name] = true
					}
				}
				for _, path := range files {
					filename := path[strings.LastIndexAny(path, "/\\")+1:]
					if childfilenamesm[filename] {
						bdel := true
						deldir := path[:strings.LastIndexAny(path, "/\\")+1]
						cnt := 0
						for name, _ := range childfilenamesm {
							//log.Println("name", parentlayerid, name)
							if toolfunc.FileExists(deldir+name) == false {
								bdel = false
								break
							} else {
								cnt += 1
							}
						}
						//log.Println(deldir, bdel, cnt)
						if bdel && cnt > 0 {
							bok := true
							for pidi := 1; pidi < parentlayerid; pidi++ {
								deldir = strings.TrimRight(deldir, "/\\")
								if strings.LastIndexAny(deldir, "/\\") != -1 {
									deldir = deldir[:strings.LastIndexAny(deldir, "/\\")+1]
								} else {
									bok = false
									break
								}
							}
							if bok {
								for dir4, _ := range hostdirsm {
									if strings.HasPrefix(deldir, dir4) {

										if ruexdirsm[deldir] == false {
											//os.RemoveAll(deldir)
											for _, rdir := range rootdirs {
												if strings.HasPrefix(path, rdir) {
													needdeldirs = append(needdeldirs, deldir[len(rdir):])
													break
												}
											}
										}

										break
									}
								}
							}
						}
					}
				}
			} else if strings.HasPrefix(li, "-pdircfds:") {
				rusub := li[len("-pdircfds:"):]
				parentlayerid := toolfunc.Atoi(rusub)
				rusub = rusub[strings.Index(rusub, ":")+1:]
				childfilenames := strings.Split(rusub, "|")
				childfilenamesm := make(map[string]bool, 0)
				for _, name := range childfilenames {
					name = strings.Trim(name, " \r\n\t")
					if name != "" {
						childfilenamesm[name] = true
					}
				}
				for _, path := range files {
					filename := path[strings.LastIndexAny(path, "/\\")+1:]
					if childfilenamesm[filename] {
						bdel := true
						cnt := 0
						deldir := path[:strings.LastIndexAny(path, "/\\")+1]
						for name, _ := range childfilenamesm {
							if !(toolfunc.FileExists(deldir+name) || toolfunc.FolderExists(deldir+name)) {
								bdel = false
								break
							} else {
								cnt += 1
							}
						}
						if bdel && cnt > 0 {
							bok := true
							for pidi := 1; pidi < parentlayerid; pidi++ {
								deldir = strings.TrimRight(deldir, "/\\")
								if strings.LastIndexAny(deldir, "/\\") != -1 {
									deldir = deldir[:strings.LastIndexAny(deldir, "/\\")+1]
								} else {
									bok = false
									break
								}
							}
							if bok {
								for dir4, _ := range hostdirsm {
									if strings.HasPrefix(deldir, dir4) {
										if ruexdirsm[deldir] == false {
											//os.RemoveAll(deldir)
											for _, rdir := range rootdirs {
												if strings.HasPrefix(path, rdir) {
													needdeldirs = append(needdeldirs, deldir[len(rdir):])
													break
												}
											}
										}

										break
									}
								}
							}
						}
					}
				}
			} else if strings.HasPrefix(li, "-dirr:") {
				re, ree := regexp.Compile(li[len("-dirr:"):])
				if ree == nil {
					for _, path := range dirs {
						ma := re.FindAllStringSubmatch(path, 1)
						if len(ma) > 0 {
							if ruexdirsm[path] == false {
								//os.RemoveAll(path)
								for _, rdir := range rootdirs {
									if strings.HasPrefix(path, rdir) {
										needdeldirs = append(needdeldirs, path[len(rdir):])
										break
									}
								}
							}

						}
					}
				}
			} else if strings.HasPrefix(li, "-direxr:") {
				sub1 := li[len("-direxr:"):]
				if strings.LastIndex(sub1, "`") == -1 {
					log.Println("error rule:", li)
					continue
				}
				re1 := sub1[:strings.LastIndex(sub1, "`")]
				repl2 := sub1[strings.LastIndex(sub1, "`")+1:]
				re, ree := regexp.Compile(re1)
				if ree == nil {
					for _, path := range dirs {
						repl := repl2
						ma := re.FindAllStringSubmatch(path, 1)
						if len(ma) > 1 {
							log.Println("direxr regex error found to match:", li, ma)
							continue
						}
						if len(ma) > 0 {
							for mai := 0; mai < len(ma[0]); mai++ {
								//fmt.Println("oldrepl", repl)
								repl = strings.ReplaceAll(repl, "$"+toolfunc.Itoa(mai), ma[0][mai])
								//fmt.Println("ReplaceAll(repl, toolfunc.Itoa(mai), ma[0][mai])", repl, "$"+toolfunc.Itoa(mai), ma[0][mai])
							}
							// for mai := 0; mai < 9; mai++ {
							// 	if strings.Index(repl, "$"+toolfunc.Itoa(mai)) != -1 {
							// 		log.Println("direxr replace with regex error:", repl, "matches", ma)
							// 	}
							// }
							var reservefiles []string
							toolfunc.GetDirAllFile(path, regexp.MustCompile(".*"), &reservefiles)
							rerv, rerve := regexp.Compile(repl)
							if rerve == nil {
								var rlrvs []string
								var rlrvsm = make(map[string]bool, 0)
								for _, rvfi := range reservefiles {
									if rerv.MatchString(rvfi) {
										rlrvs = append(rlrvs, rvfi)
										rlrvsm[rvfi] = true
									}
								}
								for _, rvfi := range reservefiles {
									if rlrvsm[rvfi] == false {
										if ruexfilesm[rvfi] == false {
											// if strings.Index(rvfi, ".") == -1 {
											// 	fmt.Println("delfile", rvfi, "repl", repl)
											// }
											//os.Remove(rvfi)
											for _, rdir := range rootdirs {
												if strings.HasPrefix(path, rdir) {
													needdelfiles = append(needdelfiles, rvfi[len(rdir):])
													break
												}
											}
										}

									}
								}
								for _, rvfi := range reservefiles {
									ffrmdir := toolfunc.GetFilePathDir(rvfi)
									if ruexdirsm[ffrmdir] == false {
										//os.Remove(ffrmdir)
										for _, rdir := range rootdirs {
											if strings.HasPrefix(path, rdir) {
												needdeldirs = append(needdeldirs, ffrmdir[len(rdir):])
												break
											}
										}
									}
								}
							} else {
								log.Println("direxr replace with regex error:", repl)
							}
						}
					}
				}
			} else if strings.HasPrefix(li, "-direqr:") {
				sub1 := li[len("-direqr:"):]
				if strings.LastIndex(sub1, "`") == -1 {
					log.Println("error rule:", li)
					continue
				}
				re1 := sub1[:strings.LastIndex(sub1, "`")]
				repl2 := sub1[strings.LastIndex(sub1, "`")+1:]
				repl2lis := strings.Split(repl2, "=match=")
				if len(repl2lis) != 2 {
					log.Println("direqr expression equaition error:", li)
				}
				re, ree := regexp.Compile(re1)
				if ree == nil {
					for _, path := range dirs {
						ma := re.FindAllStringSubmatch(path, 1)
						if len(ma) > 1 {
							log.Println("direxr regex error found to match:", li, ma)
							continue
						}
						if len(ma) > 0 {
							for mai := 0; mai < len(ma[0]); mai++ {
								repl2lis[0] = strings.ReplaceAll(repl2lis[0], "$"+toolfunc.Itoa(mai), ma[0][mai])
							}
							for mai := 0; mai < len(ma[0]); mai++ {
								repl2lis[1] = strings.ReplaceAll(repl2lis[1], "$"+toolfunc.Itoa(mai), ma[0][mai])
							}
							repre1, repre1e := regexp.Compile(repl2lis[0])
							repre2, repre2e := regexp.Compile(repl2lis[1])
							if repre1e == nil && repre1.MatchString(repl2lis[1]) || repre2e == nil && repre2.MatchString(repl2lis[0]) {
								if ruexdirsm[path] == false {
									//os.RemoveAll(path)
									for _, rdir := range rootdirs {
										if strings.HasPrefix(path, rdir) {
											needdeldirs = append(needdeldirs, path[len(rdir):])
											break
										}
									}
								}
							}
						}
					}
				}
			} else if strings.HasPrefix(li, "-filepath:") {
				fpath := li[len("-filepath:"):]

				if ruexfilesm[fpath] == false {
					//os.Remove(fpath)
					for _, rdir := range rootdirs {
						if strings.HasPrefix(fpath, rdir) {
							needdelfiles = append(needdelfiles, fpath[len(rdir):])
							break
						}
					}
				}
			} else if strings.HasPrefix(li, "-dirpath:") {
				dpath := li[len("-dirpath:"):]
				if ruexdirsm[dpath] == false {
					//os.RemoveAll(dpath)
					for _, rdir := range rootdirs {
						if strings.HasPrefix(dpath, rdir) {
							needdeldirs = append(needdeldirs, dpath[len(rdir):])
							break
						}
					}
				}
			}
		}
	}
	return needdelfiles, needdeldirs
}

func BackupFolder(srcroot, curssrcdir, tar string, bmerge bool) {
	srcsubs, subse := os.ReadDir(curssrcdir)
	if subse == nil {
		for _, subfi := range srcsubs {
			if subfi.Name() == "." || subfi.Name() == ".." {
				continue
			}
			dinfo, dinfoe := subfi.Info()
			if subfi.IsDir() {
				if needcdsm[curssrcdir+subfi.Name()] == false {
					srcdp := curssrcdir + subfi.Name()
					tardp := tar + srcdp[len(srcroot):]
					os.MkdirAll(tardp, 0666)
					BackupFolder(srcroot, curssrcdir+subfi.Name()+"/", tar, bmerge)

					if dinfoe == nil {
						os.Chtimes(tardp, dinfo.ModTime(), dinfo.ModTime())
					}

				}
			} else {
				srcfp := curssrcdir + subfi.Name()
				if needcfsm[srcfp] == false {
					tarfp := tar + srcfp[len(srcroot):]
					tarst, tarste := os.Stat(tarfp)
					// tarfdir := toolfunc.GetFilePathDir(tarfp)
					// os.MkdirAll(tarfdir)
					var tart time.Time
					if tarst != nil {
						tart, _ = time.Parse(time.RFC3339, tarst.ModTime().Format(time.RFC3339))
					}
					srct, _ := time.Parse(time.RFC3339, dinfo.ModTime().Format(time.RFC3339))
					if tarste != nil || tarste == nil && dinfoe == nil && (dinfo.Size() != tarst.Size() && tart.Unix() == srct.Unix() || tart.Unix() < srct.Unix()) {
						toolfunc.BackupFile(srcfp, tarfp)
					} else if tarste == nil && dinfoe == nil && (dinfo.Size() != tarst.Size() && tart.Unix() == srct.Unix() || tart.Unix() > srct.Unix()) {
						if bmerge {
							toolfunc.BackupFile(tarfp, srcfp)
						} else {
							targetmodgreatthansrc = append(targetmodgreatthansrc, tarfp)
						}
					}
					cpcount += 1
					if cpcount%100 == 0 {

						log.Println("BackupFile Count", cpcount, "/", totalcount)
					}
				}
			}
		}
	}
}

func main() {
	fmt.Println(`backupdir separete merge two directory command [--source= -s=] [--target= -t=] [--bimerge=true|false -bm=] [--cleantarget=true|false -ct=]`)
	if len(os.Args) == 1 {
		return
	}
	var src, tar string
	var bmerge, bclean bool
	for i, arg := range os.Args {
		if strings.Index(os.Args[i], "==") != -1 {
			panic("parameter have == error")
		}
		if strings.HasPrefix(arg, "--source=") {
			src = toolfunc.StdDir(arg[len("--source="):])
		} else if strings.HasPrefix(arg, "--target=") {
			tar = toolfunc.StdDir(arg[len("--target="):])
		} else if strings.HasPrefix(arg, "-s=") {
			src = toolfunc.StdDir(arg[len("-s="):])
		} else if strings.HasPrefix(arg, "-t=") {
			tar = toolfunc.StdDir(arg[len("-t="):])
		} else if strings.HasPrefix(arg, "--bimerge=") {
			bmerge = toolfunc.ToBool(arg[len("--bimerge="):])
		} else if strings.HasPrefix(arg, "-bm=") {
			bmerge = toolfunc.ToBool(arg[len("-bm="):])
		} else if strings.HasPrefix(arg, "--cleantarget=") {
			bclean = toolfunc.ToBool(arg[len("--cleantarget="):])
		} else if strings.HasPrefix(arg, "-ct=") {
			bclean = toolfunc.ToBool(arg[len("-ct="):])
		}
	}
	if src == "" {
		log.Println("source is empty")
		return
	}
	if tar == "" {
		log.Println("taeget is empty")
		return
	}
	log.Println("source directory", src)
	log.Println("target directory", tar)
	log.Println("clean target", bclean)
	log.Println("bimerge", bmerge)
	var fs, ds []string
	toolfunc.GetDirAllFilesAndDirs(src, &fs, &ds)
	totalcount = len(fs)
	needcfs, needcds := GetNeedCleanByRulesFiles(src+"CleanDirectoryByRules.rules", fs, ds, []string{src})
	for _, fp := range needcfs {
		needcfsm[fp] = true
	}
	for _, fp := range needcds {
		needcdsm[fp] = true
	}
	hostdirsm[src] = true
	BackupFolder(src, src, tar, bmerge)
	if bmerge {
		BackupFolder(tar, tar, src, bmerge)
	} else if bclean {
		var tfs, tds []string
		toolfunc.GetDirAllFilesAndDirs(tar, &tfs, &tds)

		for _, da := range tds {
			srcpa := src + da[len(tar):]
			if needcdsm[srcpa] == true || toolfunc.FolderExists(srcpa) == false {
				os.RemoveAll(da)
			}
		}

		for _, fa := range tfs {
			srcpa := src + fa[len(tar):]
			if needcfsm[srcpa] == true || toolfunc.FileExists(srcpa) == false {
				os.Remove(fa)
			}
		}

	}

	if len(targetmodgreatthansrc) > 0 {
		ioutil.WriteFile("targetmodgreatthansrc.log", []byte(strings.Join(targetmodgreatthansrc, "\n")), 0666)
		log.Println("目标路径文件时间大于源时间个数", len(targetmodgreatthansrc), "已经写入", "targetmodgreatthansrc.log")
	}

}
