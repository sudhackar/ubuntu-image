package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// generate README, shell completion and manpages
//go:generate go test . --generate --path ../../build/

var osExit = os.Exit

func main() {
	rootCmd := generateRootCmd()

	if err := rootCmd.Execute(); err != nil {
		osExit(1)
	}
}

func generateRootCmd() *cobra.Command {

	var rootCmd = &cobra.Command{
		Use:       "ubuntu-image",
		Short:     "Generate a bootable disk image.",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"snap", "classic"},
		Run:       func(cmd *cobra.Command, args []string) {}, // run must be non-nil to enable arg check
	}

	/*** define all the strings at the same time. This keeps the ugly
	 *** multi line strings together and leaves the rest of the code tidy ***/
	// snap specific flags
	snapFlagDesc := `Install an extra snap. This is passed through to "snap prepare-image".
The snap argument can include additional information about the channel
and/or risk with the following syntax: <snap>=<channel|risk>`
	channelFlagDesc := `The default snap channel to use`
	consoleConfFlagDesc := `Disable console-conf on the resulting image`

	// classic specific flags
	projFlagDesc := `Project name to be specified to livecd-rootfs. Mutually exclusive with --filesystem.`
	filesystemFlagDesc := `Unpacked Ubuntu filesystem to be copied to the system partition.
Mutually exclusive with --project`
	distributionFlagDesc := `Distribution name to be specified to livecd-rootfs.`
	architectureFlagDesc := `CPU architecture to be specified to livecd-rootfs. default value is 
builder arch.`
	subprojFlagDesc := `Sub project name to be specified to livecd-rootfs.`
	subarchFlagDesc := `Sub architecture to be specified to livecd-rootfs.`
	ppasFlagDesc := `Extra ppas to install. This is passed through to livecd-rootfs.`
	withProposedFlagDesc := `Proposed repo to install, This is passed through to livecd-rootfs.`

	// common flags to both commands
	sizeFlagDesc := `The suggested size of the generated disk image file. If this size is smaller than the
minimum calculated size of the image a warning will be issued and --image-size
will be ignored. The value is the size in bytes, with allowable suffixes 
'M' for MiB and 'G' for GiB. Use an extended syntax to define the suggested
size for the disk images generated by a multi-volume gadget.yaml spec.
See the ubuntu-image(1) manpage for details.`
	fileListFlagDesc := `Print to this file, a list of the file system paths to all the disk images
created by the command, if any.`
	cloudInitFlagDesc := `cloud-config data to be copied to the image`
	hooksFlagDesc := `Path or comma-separated list of paths of directories in which 
scripts for build-time hooks will be located.`
	diskInfoFlagDesc := `File to be used as .disk/info on the image's rootfs. This file can contain useful
information about the target image, like image identification data, system name,
build timestamp etc.`
	outputFlagDesc := `The directory in which to put generated disk image files.
The disk image files themselves will be named <volume>.img inside this directory,
where <volume> is the volume name taken from the gadget.yaml file.
Use this option instead of the deprecated -o/--output option.`
	debugFlagDesc := `Enable debugging output`

	// state machine flags
	directoryFlagDesc := `The working directory in which to download and unpack all the
source files for the image. This directory can exist or not, and it is not
removed after this program exits. If not given, a temporary working
directory is used instead, which *is* deleted after this program exits. Use -w
if you want to be able to resume a partial state machine run.`
	untilFlagDesc := `Run the state machine until the given STEP, non-inclusively.
STEP can be a name or number.`
	thruFlagDesc := `Run the state machine through the given STEP, inclusively.
STEP can be a name or number.`
	resumeFlagDesc := `Continue the state machine from the previously saved state.
It is an error if there is no previous state.`

	// create the snap command and add snap specific flags to it
	snap := &cobra.Command{
		Use:   "snap [model_assertion]",
		Short: "Create snap-based Ubuntu Core image.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ubuntu-image snap functionality to be added later")
		},
	}

	var Snap, Channel string
	var ConsoleConf bool
	snap.PersistentFlags().StringVarP(&Snap, "snap", "", "", snapFlagDesc)
	snap.PersistentFlags().StringVarP(&Channel, "channel", "c", "", channelFlagDesc)
	snap.PersistentFlags().BoolVarP(&ConsoleConf, "disable-console-conf", "", false, consoleConfFlagDesc)

	// create the classic command and add classic specific flags to it
	classic := &cobra.Command{
		Use:   "classic [graph_tree]",
		Short: "Create debian-based Ubuntu Classic image.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ubuntu-image classic functionality to be added later")
		},
	}

	var Project, Filesystem, Suite, Architecture, Subproject, Subarchitecture, ExtraPpas string
	var WithProposed bool

	classic.PersistentFlags().StringVarP(&Project, "project", "p", "", projFlagDesc)
	classic.PersistentFlags().StringVarP(&Filesystem, "filesystem", "f", "", filesystemFlagDesc)
	classic.PersistentFlags().StringVarP(&Suite, "suite", "s", "", distributionFlagDesc)
	classic.PersistentFlags().StringVarP(&Architecture, "arch", "a", "", architectureFlagDesc)
	classic.PersistentFlags().StringVarP(&Subproject, "subproject", "", "", subprojFlagDesc)
	classic.PersistentFlags().StringVarP(&Subarchitecture, "subarch", "", "", subarchFlagDesc)
	classic.PersistentFlags().StringVarP(&ExtraPpas, "extra-ppas", "", "", ppasFlagDesc)
	classic.PersistentFlags().BoolVarP(&WithProposed, "with-proposed", "", false, withProposedFlagDesc)

	// add common flags to both commands
	var Size, FileList, CloudInit, HooksDirectory, DiskInfo, Output string
	var Debug bool
	snap.PersistentFlags().StringVarP(&Size, "image-size", "i", "", sizeFlagDesc)
	snap.PersistentFlags().StringVarP(&FileList, "image-file-list", "", "", fileListFlagDesc)
	snap.PersistentFlags().StringVarP(&CloudInit, "cloud-init", "", "", cloudInitFlagDesc)
	snap.PersistentFlags().StringVarP(&HooksDirectory, "hooks-directory", "", "", hooksFlagDesc)
	snap.PersistentFlags().StringVarP(&DiskInfo, "disk-info", "", "", diskInfoFlagDesc)
	snap.PersistentFlags().StringVarP(&Output, "output-dir", "O", "", outputFlagDesc)
	snap.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, debugFlagDesc)

	classic.PersistentFlags().StringVarP(&Size, "image-size", "i", "", sizeFlagDesc)
	classic.PersistentFlags().StringVarP(&FileList, "image-file-list", "", "", fileListFlagDesc)
	classic.PersistentFlags().StringVarP(&CloudInit, "cloud-init", "", "", cloudInitFlagDesc)
	classic.PersistentFlags().StringVarP(&HooksDirectory, "hooks-directory", "", "", hooksFlagDesc)
	classic.PersistentFlags().StringVarP(&DiskInfo, "disk-info", "", "", diskInfoFlagDesc)
	classic.PersistentFlags().StringVarP(&Output, "output-dir", "O", "", outputFlagDesc)
	classic.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, debugFlagDesc)

	var Directory, Until, Through string
	var Resume bool

	// add state machine flags to both commands
	snap.PersistentFlags().StringVarP(&Directory, "workdir", "w", "", directoryFlagDesc)
	snap.PersistentFlags().StringVarP(&Until, "until", "u", "", untilFlagDesc)
	snap.PersistentFlags().StringVarP(&Through, "thru", "t", "", thruFlagDesc)
	snap.PersistentFlags().BoolVarP(&Resume, "resume", "r", false, resumeFlagDesc)

	classic.PersistentFlags().StringVarP(&Directory, "workdir", "w", "", directoryFlagDesc)
	classic.PersistentFlags().StringVarP(&Until, "until", "u", "", untilFlagDesc)
	classic.PersistentFlags().StringVarP(&Through, "thru", "t", "", thruFlagDesc)
	classic.PersistentFlags().BoolVarP(&Resume, "resume", "r", false, resumeFlagDesc)

	// add snap and classic to root "ubuntu-image" command
	rootCmd.AddCommand(snap)
	rootCmd.AddCommand(classic)
	return rootCmd
}
