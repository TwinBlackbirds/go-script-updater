package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var remotePath string = "" // ~/test.txt
var localPath string = ""  // text.txt
var sshPath string = "~/ssh"
var remoteUser string = "root"
var remoteIP string = ""
var preemptiveTM rune = '%'

func switchErr(err int, errstring string) {
	switch err {
	case 1:
		fmt.Println("ERROR: Exit status of 1 recieved by scp! Please ensure you are not using any non-existent directories and try again.")
		break
	case 3:
		fmt.Println("ERROR: Incorrect sshpass syntax! Please make sure that your details are correct and try again.")
		break
	case 127:
		fmt.Println("ERROR: BASH was unable to find at least one required command! Please ensure sshpass and scp are installed and properly configured, then try again.")
		break
	case 255:
		fmt.Println("ERROR: SSH was unable to establish a connection or remote path does not exist! Please ensure all filepaths are correct and port 22 is not blocked on your network, then try again.")
		break
	default:
		fmt.Printf("ERROR - UNRECOGNIZED EXIT STATUS RECIEVED: %s\n", errstring)
		break
	}
}

func validateFileExists(path string) bool {
	var fileValidCmdStr string = fmt.Sprintf("if [ -f %s ]; then echo true; else echo false; fi", path)
	output, fileValidErr := exec.Command("bash", "-c", fileValidCmdStr).Output()
	if fileValidErr != nil {
		fmt.Println("ERROR: There was an error running the following bash command!")
		fmt.Printf("CMD SENT: %s\n", fileValidCmdStr)
		fmt.Printf("RECIEVED ERROR/EXIT STATUS: %s\n", fileValidErr.Error())
		os.Exit(1)
	}
	cleanedbuf := strings.ReplaceAll(string(output), "\n", "")
	parsed, err := strconv.ParseBool(cleanedbuf)
	if err != nil {
		fmt.Println("ERROR: Error occured while trying to parse bash result as a boolean!")
		os.Exit(1)
	}
	return parsed
}

func receiver() {
	fmt.Println("Beginning update process..")
	fileExists := validateFileExists(localPath)
	if fileExists {
		println("Old version of script detected! Moving to bak file..")
		err := os.Rename(localPath, localPath+".bak")
		if err != nil {
			fmt.Printf("WARNING: There was an error (%s) backing up the script! The original may be overwritten if you continue..\n", err.Error())
			for i := 0; i < 10; i++ {
				fmt.Printf("Waiting for %d more seconds before continuing.. (CTRL+Z to exit)", 10-i)
				time.Sleep(time.Second)
			}
		}
	}
	fmt.Println("Retrieving updated files over SSH..")
	var retrievalCmd string = fmt.Sprintf("sshpass -f %s scp -P 22 %s@%s:%s %s", sshPath, remoteUser, remoteIP, remotePath, localPath)
	var sshTransferErr error = exec.Command("bash", "-c", retrievalCmd).Run()
	if sshTransferErr != nil {
		var sshTransferErrStr string = sshTransferErr.Error()
		var sshTransferErrStrSplitWS []string = strings.Split(sshTransferErrStr, " ")
		errorCode, parseErr := strconv.Atoi(sshTransferErrStrSplitWS[len(sshTransferErrStrSplitWS)-1])
		if parseErr != nil {
			fmt.Printf("ERROR: Was not able to parse error code to int! Original error was: %s\n", sshTransferErr.Error())
			os.Exit(1)
		}
		switchErr(errorCode, sshTransferErrStr)
		os.Exit(1)
	}
	fmt.Println("File retrieved successfully!")
	return
}

func uploader() {
	fmt.Println("Beginning upload process..")
	fileExists := validateFileExists(localPath)
	if !fileExists {
		fmt.Println("ERROR: Local file path does not exist! Please ensure it does and try again.")
		os.Exit(1)
	}
	var sshTransferCmdStr string = fmt.Sprintf("sshpass -f %s scp -P 22 %s %s@%s:%s", sshPath, localPath, remoteUser, remoteIP, remotePath)
	fmt.Println("Uploading files via SSH..")
	var sshTransferErr error = exec.Command("bash", "-c", sshTransferCmdStr).Run()
	if sshTransferErr != nil {
		var sshTransferErrStr string = sshTransferErr.Error()
		var sshTransferErrStrSplitWS []string = strings.Split(sshTransferErrStr, " ")
		errorCode, parseErr := strconv.Atoi(sshTransferErrStrSplitWS[len(sshTransferErrStrSplitWS)-1])
		if parseErr != nil {
			fmt.Printf("ERROR: There was an error parsing an error code to an integer! Original error: %s\n", sshTransferErr)
			os.Exit(1)
		}
		switchErr(errorCode, sshTransferErrStr)
		os.Exit(1)
	}
	fmt.Println("File transferred successfully!")
	return
}

func main() {
	fmt.Print("\n-- Welcome to Michael's Script Updater! --\n\n")
	if len(os.Args) > 1 {
		var metConditions uint8 = 0
		var amtConditionsNecessary uint8 = 5
		ip := flag.String("ip", "", "provide the ip address to use for ssh connection")
		user := flag.String("u", "", "provide the username to use for the ssh connection")
		lfp := flag.String("l", "", "provide the local filepath to use for the ssh transfer")
		rfp := flag.String("r", "", "provide the remote filepath to use for the ssh transfer")
		sshppath := flag.String("pw", "", "provide the local filepath to use for sshpass")
		transfermode := flag.String("tm", "", "preemptively provide the transfer mode. accepted values are 'u' (upload) & 'r' (recieve)")
		flag.Parse() // parse all command line flags
		if *ip != "" {
			properIP, ipErr := regexp.MatchString("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$", *ip) // simple regex match (trivial, really :p)
			if ipErr != nil {
				fmt.Println("ERROR: There was a RegEx error while trying to MatchString provided IP command line argument!")
				os.Exit(1)
			} else if !properIP {
				fmt.Println("ERROR: Invalid IP command line argument supplied!")
				os.Exit(1)
			}
			var dividedIP []string = strings.Split(*ip, ".")
			var invalidIP bool = false
			for i := 0; i < 4; i++ {
				interpretedInt, parseErr := strconv.Atoi(dividedIP[i])
				if parseErr != nil {
					fmt.Println("ERROR: Internal parsing error while attempting to parse IP command line argument as an integer!")
					os.Exit(1)
				}
				if interpretedInt > 255 {
					invalidIP = true
				}
			}
			if invalidIP {
				fmt.Println("ERROR: Invalid IP argument supplied!")
				os.Exit(1)
			}
			fmt.Println(properIP)
			remoteIP = *ip
			metConditions += 1
		}

		if *user != "" {
			remoteUser = *user
			metConditions += 1
		}
		if *lfp != "" {
			localPath = *lfp
			metConditions += 1
		}
		if *rfp != "" {
			remotePath = *rfp
			metConditions += 1
		}
		if *sshppath != "" {
			sshPath = *sshppath
			metConditions += 1
		}
		if *transfermode == "u" || *transfermode == "r" {
			preemptiveTM = rune((*transfermode)[0])
		} else if *transfermode == "" {
			// do nothing
		} else {
			fmt.Println("ERROR: Invalid commandline argument passed to transfermode (-tm)")
			os.Exit(1)
		}
		if metConditions != amtConditionsNecessary {
			fmt.Println("Some arguments not supplied! Falling back on compiled defaults..")
			// os.Exit(1)
		}
	}
	var markQuit bool = false
	if localPath == "" {
		fmt.Println("ERROR: Internal variable localPath (local filepath for file transfer) has not been set! Please edit the source and recompile. Alternatively, provide it at runtime with the command line argument '-l'")
		markQuit = true
	}
	if remotePath == "" {
		fmt.Println("ERROR: Internal variable remotePath (remote filepath for file transfer) has not been set! Please edit the source and recompile. Alternatively, provide it at runtime with the command line argument '-r'")
		markQuit = true
	}
	if sshPath == "" {
		fmt.Println("ERROR: Internal variable sshPath (filepath to be passed to sshpass) has not been set! Please edit the source and recompile. Alternatively, provide it at runtime with the command line argument '-pw'")
		markQuit = true
	}
	if remoteIP == "" {
		fmt.Println("ERROR: Internal variable remoteIP (remote server ip for ssh) has not been set! Please edit the source and recompile. Alternatively, provide it at runtime with the command line argument '-ip'")
		markQuit = true
	}
	if remoteUser == "" {
		fmt.Println("ERROR: Internal variable remoteUser (remote server login user for ssh) have not been set! Please edit the source and recompile. Alternatively, provide it at runtime with the command line argument '-u'")
		markQuit = true
	}
	if markQuit {
		os.Exit(1)
	}
	if preemptiveTM != '%' {
		fmt.Println("Preemptive transfermode value detected! Proceeding promptless..")
		if preemptiveTM == 'u' {
			uploader()
		} else if preemptiveTM == 'r' {
			receiver()
		}
	} else {
		fmt.Printf("Would you like to upload or receive? (u/r) ")
		var buf string
		fmt.Scanf("%s", &buf)
		var input string = strings.ReplaceAll(buf, "\n", "")
		if len(input) == 0 {
			fmt.Println("Invalid response!")
			return
		}
		var lower rune = rune(strings.ToLower(input)[0])
		if lower == 'u' {
			fmt.Println("Uploading..")
			uploader()
		} else if lower == 'r' {
			fmt.Println("Recieving..")
			receiver()
		} else {
			fmt.Println("Invalid response!")
		}
	}
}
