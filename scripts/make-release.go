/*
   Network Next. Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runCommand(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		fmt.Printf("runCommand error: %v\n", err)
		return false
	}
	return true
}

func bash(format string, args ...interface{}) bool {
	command := fmt.Sprintf(format, args...)
	return runCommand("bash", []string{"-c", command})
}

func main() {

	version := "3.4.4"

	basedir := fmt.Sprintf("next-%s", version)
	
	fmt.Printf("\nMaking release %s in \"%s\"\n", version, basedir)

	// create the release directory clean
	bash("rm -rf %s", basedir)
	bash("mkdir -p %s", basedir)

	// create the include dir and copy across next.h
	bash("mkdir -p %s/include", basedir)
	bash("cp include/next.h %s/include", basedir)

	// copy across the debug win64 libraries
	libdir := "lib/Windows-x86_64/Debug"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.lib %s/%s/next-win64-debug-%s.lib", libdir, basedir, libdir, version)
	bash("cp %s/next.pdb %s/%s/next-win64-debug-%s.pdb", libdir, basedir, libdir, version)
	bash("cp %s/next.dll %s/%s/next-win64-debug-%s.dll", libdir, basedir, libdir, version)
	bash("cp %s/next.exp %s/%s/next-win64-debug-%s.exp", libdir, basedir, libdir, version)

	// copy across the release win64 libraries
	libdir = "lib/Windows-x86_64/Release"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.lib %s/%s/next-win64-release-%s.lib", libdir, basedir, libdir, version)
	bash("cp %s/next.pdb %s/%s/next-win64-release-%s.pdb", libdir, basedir, libdir, version)
	bash("cp %s/next.dll %s/%s/next-win64-release-%s.dll", libdir, basedir, libdir, version)
	bash("cp %s/next.exp %s/%s/next-win64-release-%s.exp", libdir, basedir, libdir, version)

	// copy across the debug win32 libraries
	libdir = "lib/Windows-x86/Debug"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.lib %s/%s/next-win32-debug-%s.lib", libdir, basedir, libdir, version)
	bash("cp %s/next.pdb %s/%s/next-win32-debug-%s.pdb", libdir, basedir, libdir, version)
	bash("cp %s/next.dll %s/%s/next-win32-debug-%s.dll", libdir, basedir, libdir, version)
	bash("cp %s/next.exp %s/%s/next-win32-debug-%s.exp", libdir, basedir, libdir, version)

	// copy across the release win32 libraries
	libdir = "lib/Windows-x86/Release"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.lib %s/%s/next-win32-release-%s.lib", libdir, basedir, libdir, version)
	bash("cp %s/next.pdb %s/%s/next-win32-release-%s.pdb", libdir, basedir, libdir, version)
	bash("cp %s/next.dll %s/%s/next-win32-release-%s.dll", libdir, basedir, libdir, version)
	bash("cp %s/next.exp %s/%s/next-win32-release-%s.exp", libdir, basedir, libdir, version)

	// copy across the debug x1 libraries
	libdir = "lib/XBoxOne/Debug"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.lib %s/%s/next-x1-debug-%s.lib", libdir, basedir, libdir, version)
	bash("cp %s/next.pdb %s/%s/next-x1-debug-%s.pdb", libdir, basedir, libdir, version)
	bash("cp %s/next.dll %s/%s/next-x1-debug-%s.dll", libdir, basedir, libdir, version)
	bash("cp %s/next.exp %s/%s/next-x1-debug-%s.exp", libdir, basedir, libdir, version)

	// copy across the release x1 libraries
	libdir = "lib/XBoxOne/Release"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.lib %s/%s/next-x1-release-%s.lib", libdir, basedir, libdir, version)
	bash("cp %s/next.pdb %s/%s/next-x1-release-%s.pdb", libdir, basedir, libdir, version)
	bash("cp %s/next.dll %s/%s/next-x1-release-%s.dll", libdir, basedir, libdir, version)
	bash("cp %s/next.exp %s/%s/next-x1-release-%s.exp", libdir, basedir, libdir, version)

	// copy across the debug ps4 libraries
	libdir = "lib/Playstation4/Debug"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next_stub_weak.a %s/%s/next-ps4-debug-%s_stub_weak.a", libdir, basedir, libdir, version)
	bash("cp %s/next_stub.a %s/%s/next-ps4-debug-%s_stub.a", libdir, basedir, libdir, version)
	bash("cp %s/next.prx %s/%s/next-ps4-debug-%s.prx", libdir, basedir, libdir, version)

	// copy across the release ps4 libraries
	libdir = "lib/Playstation4/Release"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next_stub_weak.a %s/%s/next-ps4-release-%s_stub_weak.a", libdir, basedir, libdir, version)
	bash("cp %s/next_stub.a %s/%s/next-ps4-release-%s_stub.a", libdir, basedir, libdir, version)
	bash("cp %s/next.prx %s/%s/next-ps4-release-%s.prx", libdir, basedir, libdir, version)

	// copy across the debug nx64 libraries
	libdir = "lib/NintendoSwitch-NX64/Debug"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.nro %s/%s/next-nx64-debug-%s.nro", libdir, basedir, libdir, version)
	bash("cp %s/next.nrr %s/%s/next-nx64-debug-%s.nrr", libdir, basedir, libdir, version)
	bash("cp %s/next.nrs %s/%s/next-nx64-debug-%s.nrs", libdir, basedir, libdir, version)

	// copy across the release nx64 libraries
	libdir = "lib/NintendoSwitch-NX64/Release"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.nro %s/%s/next-nx64-release-%s.nro", libdir, basedir, libdir, version)
	bash("cp %s/next.nrr %s/%s/next-nx64-release-%s.nrr", libdir, basedir, libdir, version)
	bash("cp %s/next.nrs %s/%s/next-nx64-release-%s.nrs", libdir, basedir, libdir, version)

	// copy across the debug nx32 libraries
	libdir = "lib/NintendoSwitch-NX32/Debug"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.nro %s/%s/next-nx32-debug-%s.nro", libdir, basedir, libdir, version)
	bash("cp %s/next.nrr %s/%s/next-nx32-debug-%s.nrr", libdir, basedir, libdir, version)
	bash("cp %s/next.nrs %s/%s/next-nx32-debug-%s.nrs", libdir, basedir, libdir, version)

	// copy across the release nx32 libraries
	libdir = "lib/NintendoSwitch-NX32/Release"
	bash("mkdir -p %s/%s", basedir, libdir)
	bash("cp %s/next.nro %s/%s/next-nx32-release-%s.nro", libdir, basedir, libdir, version)
	bash("cp %s/next.nrr %s/%s/next-nx32-release-%s.nrr", libdir, basedir, libdir, version)
	bash("cp %s/next.nrs %s/%s/next-nx32-release-%s.nrs", libdir, basedir, libdir, version)

	// build manifest

	manifest := make([]string, 0)

	manifest = append(manifest, fmt.Sprintf("%s/include/next.h", basedir))

	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Debug/next-win64-debug-%s.lib", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Debug/next-win64-debug-%s.pdb", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Debug/next-win64-debug-%s.dll", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Debug/next-win64-debug-%s.exp", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Release/next-win64-release-%s.lib", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Release/next-win64-release-%s.pdb", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Release/next-win64-release-%s.dll", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86_64/Release/next-win64-release-%s.exp", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Debug/next-win32-debug-%s.lib", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Debug/next-win32-debug-%s.pdb", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Debug/next-win32-debug-%s.dll", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Debug/next-win32-debug-%s.exp", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Release/next-win32-release-%s.lib", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Release/next-win32-release-%s.pdb", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Release/next-win32-release-%s.dll", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Windows-x86/Release/next-win32-release-%s.exp", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Debug/next-x1-debug-%s.lib", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Debug/next-x1-debug-%s.pdb", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Debug/next-x1-debug-%s.dll", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Debug/next-x1-debug-%s.exp", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Release/next-x1-release-%s.lib", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Release/next-x1-release-%s.pdb", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Release/next-x1-release-%s.dll", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/XBoxOne/Release/next-x1-release-%s.exp", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/Playstation4/Debug/next-ps4-debug-%s_stub_weak.a", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Playstation4/Debug/next-ps4-debug-%s_stub.a", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Playstation4/Debug/next-ps4-debug-%s.prx", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/Playstation4/Release/next-ps4-release-%s_stub_weak.a", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Playstation4/Release/next-ps4-release-%s_stub.a", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/Playstation4/Release/next-ps4-release-%s.prx", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX64/Debug/next-nx64-debug-%s.nro", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX64/Debug/next-nx64-debug-%s.nrr", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX64/Debug/next-nx64-debug-%s.nrs", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX64/Release/next-nx64-release-%s.nro", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX64/Release/next-nx64-release-%s.nrr", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX64/Release/next-nx64-release-%s.nrs", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX32/Debug/next-nx32-debug-%s.nro", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX32/Debug/next-nx32-debug-%s.nrr", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX32/Debug/next-nx32-debug-%s.nrs", basedir, version))

	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX32/Release/next-nx32-release-%s.nro", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX32/Release/next-nx32-release-%s.nrr", basedir, version))
	manifest = append(manifest, fmt.Sprintf("%s/lib/NintendoSwitch-NX32/Release/next-nx32-release-%s.nrs", basedir, version))

	fmt.Printf("\nManifest:\n\n" )
	for _, file := range manifest {
		fmt.Printf("    %s\n", file)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("\n    ^---- error: missing file!\n\n")
			os.Exit(1)
		}
	}

	fmt.Printf("\nSuccessfully made release %s\n\n", version)

}
