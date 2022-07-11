package statemachine

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/canonical/ubuntu-image/internal/helper"
	"github.com/google/uuid"
	"github.com/snapcore/snapd/gadget"
	"github.com/snapcore/snapd/gadget/quantity"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/osutil/mkfs"
)

// TestMaxOffset tests the functionality of the maxOffset function
func TestMaxOffset(t *testing.T) {
	t.Run("test_max_offset", func(t *testing.T) {
		lesser := quantity.Offset(0)
		greater := quantity.Offset(1)

		if maxOffset(lesser, greater) != greater {
			t.Errorf("maxOffset returned the lower number")
		}

		// reverse argument order
		if maxOffset(greater, lesser) != greater {
			t.Errorf("maxOffset returned the lower number")
		}
	})
}

// TestFailedRunHooks tests failures in the runHooks function. This is accomplished by mocking
// functions and calling hook scripts that intentionally return errors
func TestFailedRunHooks(t *testing.T) {
	t.Run("test_failed_run_hooks", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		stateMachine.commonFlags.Debug = true // for coverage!

		// need workdir set up for this
		err := stateMachine.makeTemporaryDirectories()
		asserter.AssertErrNil(err, true)

		// first set a good hooks directory
		stateMachine.commonFlags.HooksDirectories = []string{filepath.Join(
			"testdata", "good_hookscript")}
		// mock ioutil.ReadDir
		ioutilReadDir = mockReadDir
		defer func() {
			ioutilReadDir = ioutil.ReadDir
		}()
		err = stateMachine.runHooks("post-populate-rootfs",
			"UBUNTU_IMAGE_HOOK_ROOTFS", stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error reading hooks directory")
		ioutilReadDir = ioutil.ReadDir

		// now set a hooks directory that will fail
		stateMachine.commonFlags.HooksDirectories = []string{filepath.Join(
			"testdata", "hooks_return_error")}
		err = stateMachine.runHooks("post-populate-rootfs",
			"UBUNTU_IMAGE_HOOK_ROOTFS", stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error running hook")
		os.RemoveAll(stateMachine.stateMachineFlags.WorkDir)
	})
}

// TestFailedHandleSecureBoot tests failures in the handleSecureBoot function by mocking functions
func TestFailedHandleSecureBoot(t *testing.T) {
	t.Run("test_failed_handle_secure_boot", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()

		// need workdir for this
		if err := stateMachine.makeTemporaryDirectories(); err != nil {
			t.Errorf("Did not expect an error, got %s", err.Error())
		}

		// create a volume
		volume := new(gadget.Volume)
		volume.Bootloader = "u-boot"
		// make the u-boot directory and add a file
		bootDir := filepath.Join(stateMachine.tempDirs.unpack,
			"image", "boot", "uboot")
		os.MkdirAll(bootDir, 0755)
		osutil.CopySpecialFile(filepath.Join("testdata", "grubenv"), bootDir)

		// mock os.Mkdir
		osMkdirAll = mockMkdirAll
		defer func() {
			osMkdirAll = os.MkdirAll
		}()
		err := stateMachine.handleSecureBoot(volume, stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error creating ubuntu dir")
		osMkdirAll = os.MkdirAll

		// mock ioutil.ReadDir
		ioutilReadDir = mockReadDir
		defer func() {
			ioutilReadDir = ioutil.ReadDir
		}()
		err = stateMachine.handleSecureBoot(volume, stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error reading boot dir")
		ioutilReadDir = ioutil.ReadDir

		// mock os.Rename
		osRename = mockRename
		defer func() {
			osRename = os.Rename
		}()
		err = stateMachine.handleSecureBoot(volume, stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error copying boot dir")
		osRename = os.Rename
	})
}

// TestFailedHandleSecureBootPiboot tests failures in the handleSecureBoot
// function by mocking functions, for piboot
func TestFailedHandleSecureBootPiboot(t *testing.T) {
	t.Run("test_failed_handle_secure_boot_piboot", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()

		// need workdir for this
		if err := stateMachine.makeTemporaryDirectories(); err != nil {
			t.Errorf("Did not expect an error, got %s", err.Error())
		}

		// create a volume
		volume := new(gadget.Volume)
		volume.Bootloader = "piboot"
		// make the piboot directory and add a file
		bootDir := filepath.Join(stateMachine.tempDirs.unpack,
			"image", "boot", "piboot")
		os.MkdirAll(bootDir, 0755)
		osutil.CopySpecialFile(filepath.Join("testdata", "gadget_tree_piboot",
			"piboot.conf"), bootDir)

		// mock os.Mkdir
		osMkdirAll = mockMkdirAll
		defer func() {
			osMkdirAll = os.MkdirAll
		}()
		err := stateMachine.handleSecureBoot(volume, stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error creating ubuntu dir")
		osMkdirAll = os.MkdirAll

		// mock ioutil.ReadDir
		ioutilReadDir = mockReadDir
		defer func() {
			ioutilReadDir = ioutil.ReadDir
		}()
		err = stateMachine.handleSecureBoot(volume, stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error reading boot dir")
		ioutilReadDir = ioutil.ReadDir

		// mock os.Rename
		osRename = mockRename
		defer func() {
			osRename = os.Rename
		}()
		err = stateMachine.handleSecureBoot(volume, stateMachine.tempDirs.rootfs)
		asserter.AssertErrContains(err, "Error copying boot dir")
		osRename = os.Rename
	})
}

// TestHandleLkBootloader tests that the handleLkBootloader function runs successfully
func TestHandleLkBootloader(t *testing.T) {
	t.Run("test_handle_lk_bootloader", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		stateMachine.YamlFilePath = filepath.Join("testdata", "gadget_tree",
			"meta", "gadget.yaml")

		// need workdir set up for this
		err := stateMachine.makeTemporaryDirectories()
		asserter.AssertErrNil(err, true)

		// create image/boot/lk and place a test file there
		bootDir := filepath.Join(stateMachine.tempDirs.unpack, "image", "boot", "lk")
		err = os.MkdirAll(bootDir, 0755)
		asserter.AssertErrNil(err, true)

		err = osutil.CopyFile(filepath.Join("testdata", "disk_info"),
			filepath.Join(bootDir, "disk_info"), 0)
		asserter.AssertErrNil(err, true)

		// set up the volume
		volume := new(gadget.Volume)
		volume.Bootloader = "lk"

		err = stateMachine.handleLkBootloader(volume)
		asserter.AssertErrNil(err, true)

		// ensure the test file was moved
		movedFile := filepath.Join(stateMachine.tempDirs.unpack, "gadget", "disk_info")
		if _, err := os.Stat(movedFile); err != nil {
			t.Errorf("File %s should exist but it does not", movedFile)
		}
	})
}

// TestFailedHandleLkBootloader tests failures in handleLkBootloader by mocking functions
func TestFailedHandleLkBootloader(t *testing.T) {
	t.Run("test_failed_handle_lk_bootloader", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		stateMachine.YamlFilePath = filepath.Join("testdata", "gadget_tree",
			"meta", "gadget.yaml")

		// need workdir set up for this
		err := stateMachine.makeTemporaryDirectories()
		asserter.AssertErrNil(err, true)
		// create image/boot/lk and place a test file there
		bootDir := filepath.Join(stateMachine.tempDirs.unpack, "image", "boot", "lk")
		err = os.MkdirAll(bootDir, 0755)
		asserter.AssertErrNil(err, true)

		err = osutil.CopyFile(filepath.Join("testdata", "disk_info"),
			filepath.Join(bootDir, "disk_info"), 0)
		asserter.AssertErrNil(err, true)

		// set up the volume
		volume := new(gadget.Volume)
		volume.Bootloader = "lk"

		// mock os.Mkdir
		osMkdir = mockMkdir
		defer func() {
			osMkdir = os.Mkdir
		}()
		err = stateMachine.handleLkBootloader(volume)
		asserter.AssertErrContains(err, "Failed to create gadget dir")
		osMkdir = os.Mkdir

		// mock ioutil.ReadDir
		ioutilReadDir = mockReadDir
		defer func() {
			ioutilReadDir = ioutil.ReadDir
		}()
		err = stateMachine.handleLkBootloader(volume)
		asserter.AssertErrContains(err, "Error reading lk bootloader dir")
		ioutilReadDir = ioutil.ReadDir

		// mock osutil.CopySpecialFile
		osutilCopySpecialFile = mockCopySpecialFile
		defer func() {
			osutilCopySpecialFile = osutil.CopySpecialFile
		}()
		err = stateMachine.handleLkBootloader(volume)
		asserter.AssertErrContains(err, "Error copying lk bootloader dir")
		osutilCopySpecialFile = osutil.CopySpecialFile
	})
}

// TestFailedCopyStructureContent tests failures in the copyStructureContent function by mocking
// functions and setting invalid bs= arguments in dd
func TestFailedCopyStructureContent(t *testing.T) {
	t.Run("test_failed_copy_structure_content", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		stateMachine.YamlFilePath = filepath.Join("testdata", "gadget_tree",
			"meta", "gadget.yaml")

		// need workdir and loaded gadget.yaml set up for this
		err := stateMachine.makeTemporaryDirectories()
		asserter.AssertErrNil(err, true)
		err = stateMachine.loadGadgetYaml()
		asserter.AssertErrNil(err, true)

		// separate out the volumeStructures to test different scenarios
		var mbrStruct gadget.VolumeStructure
		var rootfsStruct gadget.VolumeStructure
		var volume *gadget.Volume = stateMachine.GadgetInfo.Volumes["pc"]
		for _, structure := range volume.Structure {
			if structure.Name == "mbr" {
				mbrStruct = structure
			} else if structure.Name == "EFI System" {
				rootfsStruct = structure
			}
		}

		// mock helper.CopyBlob and test with no filesystem specified
		helperCopyBlob = mockCopyBlob
		defer func() {
			helperCopyBlob = helper.CopyBlob
		}()
		err = stateMachine.copyStructureContent(volume, mbrStruct, 0, "",
			filepath.Join("/tmp", uuid.NewString()+".img"))
		asserter.AssertErrContains(err, "Error zeroing partition")
		helperCopyBlob = helper.CopyBlob

		// set an invalid blocksize to mock the binary copy blob
		mockableBlockSize = "0"
		defer func() {
			mockableBlockSize = "1"
		}()
		err = stateMachine.copyStructureContent(volume, mbrStruct, 0, "",
			filepath.Join("/tmp", uuid.NewString()+".img"))
		asserter.AssertErrContains(err, "Error copying image blob")
		mockableBlockSize = "1"

		// mock helper.CopyBlob and test with filesystem: vfat
		helperCopyBlob = mockCopyBlob
		defer func() {
			helperCopyBlob = helper.CopyBlob
		}()
		err = stateMachine.copyStructureContent(volume, rootfsStruct, 0, "",
			filepath.Join("/tmp", uuid.NewString()+".img"))
		asserter.AssertErrContains(err, "Error zeroing image file")
		helperCopyBlob = helper.CopyBlob

		// mock gadget.MkfsWithContent
		mkfsMakeWithContent = mockMkfsWithContent
		defer func() {
			mkfsMakeWithContent = mkfs.MakeWithContent
		}()
		err = stateMachine.copyStructureContent(volume, rootfsStruct, 0, "",
			filepath.Join("/tmp", uuid.NewString()+".img"))
		asserter.AssertErrContains(err, "Error running mkfs with content")
		mkfsMakeWithContent = mkfs.MakeWithContent

		// mock mkfs.Mkfs
		rootfsStruct.Content = nil // to trigger the "empty partition" case
		mkfsMake = mockMkfs
		defer func() {
			mkfsMake = mkfs.Make
		}()
		err = stateMachine.copyStructureContent(volume, rootfsStruct, 0, "",
			filepath.Join("/tmp", uuid.NewString()+".img"))
		asserter.AssertErrContains(err, "Error running mkfs")
		mkfsMake = mkfs.Make
	})
}

// TestCleanup ensures that the temporary workdir is cleaned up after the
// state machine has finished running
func TestCleanup(t *testing.T) {
	t.Run("test_cleanup", func(t *testing.T) {
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		stateMachine.Run()
		stateMachine.Teardown()
		if _, err := os.Stat(stateMachine.stateMachineFlags.WorkDir); err == nil {
			t.Errorf("Error: temporary workdir %s was not cleaned up\n",
				stateMachine.stateMachineFlags.WorkDir)
		}
	})
}

// TestFailedCleanup tests a failure in os.RemoveAll while deleting the temporary directory
func TestFailedCleanup(t *testing.T) {
	t.Run("test_failed_cleanup", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		stateMachine.cleanWorkDir = true

		osRemoveAll = mockRemoveAll
		defer func() {
			osRemoveAll = os.RemoveAll
		}()
		err := stateMachine.cleanup()
		asserter.AssertErrContains(err, "Error cleaning up workDir")
	})
}

// TestFailedCalculateImageSize tests a scenario when calculateImageSize() is called
// with a nil pointer to stateMachine.GadgetInfo
func TestFailedCalculateImageSize(t *testing.T) {
	t.Run("test_failed_calculate_image_size", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		_, err := stateMachine.calculateImageSize()
		asserter.AssertErrContains(err, "Cannot calculate image size before initializing GadgetInfo")
	})
}

// TestFailedWriteOffsetValues tests various error scenarios for writeOffsetValues
func TestFailedWriteOffsetValues(t *testing.T) {
	t.Run("test_failed_write_offset_values", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
		stateMachine.YamlFilePath = filepath.Join("testdata", "gadget_tree",
			"meta", "gadget.yaml")

		// need workdir and loaded gadget.yaml set up for this
		err := stateMachine.makeTemporaryDirectories()
		asserter.AssertErrNil(err, true)
		err = stateMachine.loadGadgetYaml()
		asserter.AssertErrNil(err, true)

		// create an empty pc.img
		imgPath := filepath.Join(stateMachine.stateMachineFlags.WorkDir, "pc.img")
		os.Create(imgPath)
		os.Truncate(imgPath, 0)

		volume, found := stateMachine.GadgetInfo.Volumes["pc"]
		if !found {
			t.Fatalf("Failed to find gadget volume")
		}
		// pass an image size that's too small
		err = writeOffsetValues(volume, imgPath, 512, 4)
		asserter.AssertErrContains(err, "write offset beyond end of file")

		// mock os.Open file to force it to use os.O_APPEND, which causes
		// errors in file.WriteAt()
		osOpenFile = mockOpenFileAppend
		defer func() {
			osOpenFile = os.OpenFile
		}()
		err = writeOffsetValues(volume, imgPath, 512, 0)
		asserter.AssertErrContains(err, "Failed to write offset to disk")
		osOpenFile = os.OpenFile
	})
}

// TestWarningRootfsSizeTooSmall tests that a warning is thrown if the structure size
// for the rootfs specified in gadget.yaml is smaller than the calculated rootfs size.
// It also ensures that the size is corrected in the structure struct
func TestWarningRootfsSizeTooSmall(t *testing.T) {
	t.Run("test_warning_rootfs_size_too_small", func(t *testing.T) {
		asserter := helper.Asserter{T: t}
		var stateMachine StateMachine
		stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()

		stateMachine.YamlFilePath = filepath.Join("testdata", "gadget_tree",
			"meta", "gadget.yaml")

		// need workdir and loaded gadget.yaml set up for this
		err := stateMachine.makeTemporaryDirectories()
		asserter.AssertErrNil(err, true)
		err = stateMachine.loadGadgetYaml()
		asserter.AssertErrNil(err, true)

		// set up a "rootfs" that we can calculate the size of
		os.MkdirAll(stateMachine.tempDirs.rootfs, 0755)
		osutil.CopySpecialFile(filepath.Join("testdata", "gadget_tree"), stateMachine.tempDirs.rootfs)

		// ensure volumes exists
		os.MkdirAll(stateMachine.tempDirs.volumes, 0755)

		// calculate the size of the rootfs
		err = stateMachine.calculateRootfsSize()
		asserter.AssertErrNil(err, true)

		// manually set the size of the rootfs structure to 0
		var volume *gadget.Volume = stateMachine.GadgetInfo.Volumes["pc"]
		var rootfsStructure gadget.VolumeStructure
		var rootfsStructureNumber int
		for structureNumber, structure := range volume.Structure {
			if structure.Role == gadget.SystemData {
				structure.Size = 0
				rootfsStructure = structure
				rootfsStructureNumber = structureNumber
			}
		}

		// capture stdout, run copy structure content, and ensure the warning was thrown
		stdout, restoreStdout, err := helper.CaptureStd(&os.Stdout)
		defer restoreStdout()
		asserter.AssertErrNil(err, true)

		err = stateMachine.copyStructureContent(volume,
			rootfsStructure,
			rootfsStructureNumber,
			stateMachine.tempDirs.rootfs,
			filepath.Join(stateMachine.tempDirs.volumes, "part0.img"))
		asserter.AssertErrNil(err, true)

		// restore stdout and check that the warning was printed
		restoreStdout()
		readStdout, err := ioutil.ReadAll(stdout)
		asserter.AssertErrNil(err, true)

		if !strings.Contains(string(readStdout), "WARNING: rootfs structure size 0 B smaller than actual rootfs contents") {
			t.Errorf("Warning about structure size to small not present in stdout: \"%s\"", string(readStdout))
		}

		// check that the size was correctly updated in the volume
		for _, structure := range volume.Structure {
			if structure.Role == gadget.SystemData {
				if structure.Size != stateMachine.RootfsSize {
					t.Errorf("rootfs structure size %s is not equal to calculated size %s",
						structure.Size.IECString(),
						stateMachine.RootfsSize.IECString())
				}
			}
		}
	})
}

// TestGetStructureOffset ensures structure offset safely dereferences structure.Offset
func TestGetStructureOffset(t *testing.T) {
	var testOffset quantity.Offset = 1
	testCases := []struct {
		name      string
		structure gadget.VolumeStructure
		expected  quantity.Offset
	}{
		{"nil", gadget.VolumeStructure{Offset: nil}, 0},
		{"with_value", gadget.VolumeStructure{Offset: &testOffset}, 1},
	}
	for _, tc := range testCases {
		t.Run("test_get_structure_offset_"+tc.name, func(t *testing.T) {
			offset := getStructureOffset(tc.structure)
			if offset != tc.expected {
				t.Errorf("Error, expected offset %d but got %d", tc.expected, offset)
			}
		})
	}
}

// TestGenerateUniqueDiskID ensures that we generate unique disk IDs
func TestGenerateUniqueDiskID(t *testing.T) {
	testCases := []struct {
		name        string
		existing    [][]byte
		randomBytes [][]byte
		expected    []byte
		expectedErr bool
	}{
		{"one_time", [][]byte{{4, 5, 6, 7}}, [][]byte{{0, 1, 2, 3}}, []byte{0, 1, 2, 3}, false},
		{"collision", [][]byte{{0, 1, 2, 3}}, [][]byte{{0, 1, 2, 3}, {4, 5, 6, 7}}, []byte{4, 5, 6, 7}, false},
		{"broken", [][]byte{{0, 0, 0, 0}}, nil, []byte{0, 0, 0, 0}, true},
	}
	for _, tc := range testCases {
		t.Run("test_generate_unique_diskid_"+tc.name, func(t *testing.T) {
			asserter := helper.Asserter{T: t}
			// create a test rng reader, using data from our testcase
			ithRead := 0
			randRead = func(output []byte) (int, error) {
				var randomBytes []byte
				if tc.randomBytes == nil || ithRead > (len(tc.randomBytes)-1) {
					randomBytes = []byte{0, 0, 0, 0}
				} else {
					randomBytes = tc.randomBytes[ithRead]
				}
				copy(output, randomBytes)
				ithRead++
				return 0, nil
			}
			defer func() {
				randRead = rand.Read
			}()

			randomBytes, err := generateUniqueDiskID(&tc.existing)
			if tc.expectedErr {
				asserter.AssertErrContains(err, "Failed to generate unique disk ID")
			} else {
				asserter.AssertErrNil(err, true)
				if bytes.Compare(randomBytes, tc.expected) != 0 {
					t.Errorf("Error, expected ID %v but got %v", tc.expected, randomBytes)
				}
				// check if the ID was added to the list of existing IDs
				found := false
				for _, id := range tc.existing {
					if bytes.Compare(id, randomBytes) == 0 {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Error, disk ID not added to the existing list")
				}
			}
		})
	}
}

// TestGetHostArch unit tests the getHostArch function
func TestGetHostArch(t *testing.T) {
	t.Run("test_get_host_arch", func(t *testing.T) {
		hostArch := getHostArch()
		switch runtime.GOARCH {
		case "amd64":
			expected := "amd64"
			if hostArch != expected {
				t.Errorf("Wrong value of getHostArch. Expected %s, got %s", expected, hostArch)
			}
			break
		case "arm":
			expected := "armhf"
			if hostArch != expected {
				t.Errorf("Wrong value of getHostArch. Expected %s, got %s", "amd64", hostArch)
			}
			break
		case "arm64":
			expected := "arm64"
			if hostArch != expected {
				t.Errorf("Wrong value of getHostArch. Expected %s, got %s", "amd64", hostArch)
			}
			break
		case "ppc64le":
			expected := "ppc64el"
			if hostArch != expected {
				t.Errorf("Wrong value of getHostArch. Expected %s, got %s", "amd64", hostArch)
			}
			break
		case "s390x":
			expected := "s390x"
			if hostArch != expected {
				t.Errorf("Wrong value of getHostArch. Expected %s, got %s", "amd64", hostArch)
			}
			break
		case "riscv64":
			expected := "riscv64"
			if hostArch != expected {
				t.Errorf("Wrong value of getHostArch. Expected %s, got %s", "amd64", hostArch)
			}
			break
		default:
			t.Skipf("Test not supported on architecture %s", runtime.GOARCH)
			break
		}
	})
}

// TestGetHostSuite unit tests the getHostSuite function to make sure
// it returns a string with length greater than zero
func TestGetHostSuite(t *testing.T) {
	t.Run("test_get_host_suite", func(t *testing.T) {
		hostSuite := getHostSuite()
		if len(hostSuite) == 0 {
			t.Error("getHostSuite could not get the host suite")
		}
	})
}

// TestGetQemuStaticForArch unit tests the getQemuStaticForArch function
func TestGetQemuStaticForArch(t *testing.T) {
	testCases := []struct {
		arch     string
		expected string
	}{
		{"amd64", ""},
		{"armhf", "qemu-arm-static"},
		{"arm64", "qemu-aarch64-static"},
		{"ppc64el", "qemu-ppc64le-static"},
		{"s390x", ""},
		{"riscv64", ""},
	}
	for _, tc := range testCases {
		t.Run("test_get_qemu_static_for_"+tc.arch, func(t *testing.T) {
			qemuStatic := getQemuStaticForArch(tc.arch)
			if qemuStatic != tc.expected {
				t.Errorf("Expected qemu static \"%s\" for arch \"%s\", instead got \"%s\"",
					tc.expected, tc.arch, qemuStatic)
			}
		})
	}
}

// TestGenerateGerminateCmd unit tests the generateGerminateCmd function
func TestGenerateGerminateCmd(t *testing.T) {
	testCases := []struct {
		name   string
		mirror string
	}{
		{"amd64", "http://archive.ubuntu.com/ubuntu/"},
		{"armhf", "http://ports.ubuntu.com/ubuntu/"},
		{"arm64", "http://ports.ubuntu.com/ubuntu/"},
	}
	for _, tc := range testCases {
		t.Run("test_generate_germinate_cmd_"+tc.name, func(t *testing.T) {
			imageDef := ImageDefinition{
				Architecture: tc.name,
				Rootfs: &RootfsType{
					Mirror: tc.mirror,
					Seed: &SeedType{
						SeedURLs:   []string{"git://test.git"},
						SeedBranch: "testbranch",
					},
					Components: []string{"main", "universe"},
				},
			}
			germinateCmd := generateGerminateCmd(imageDef)

			if !strings.Contains(germinateCmd.String(), tc.mirror) {
				t.Errorf("germinate command \"%s\" has incorrect mirror. Expected \"%s\"",
					germinateCmd.String(), tc.mirror)
			}

			if !strings.Contains(germinateCmd.String(), "--components=main,universe") {
				t.Errorf("Expected germinate command \"%s\" to contain "+
					"\"--components=main,universe\"", germinateCmd.String())
			}
		})
	}
}

// TestValidateInput tests that invalid state machine command line arguments result in a failure
func TestValidateInput(t *testing.T) {
	testCases := []struct {
		name   string
		until  string
		thru   string
		resume bool
		errMsg string
	}{
		{"both_until_and_thru", "make_temporary_directories", "calculate_rootfs_size", false, "cannot specify both --until and --thru"},
		{"resume_with_no_workdir", "", "", true, "must specify workdir when using --resume flag"},
	}
	for _, tc := range testCases {
		t.Run("test "+tc.name, func(t *testing.T) {
			asserter := helper.Asserter{T: t}
			var stateMachine StateMachine
			stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
			stateMachine.stateMachineFlags.Until = tc.until
			stateMachine.stateMachineFlags.Thru = tc.thru
			stateMachine.stateMachineFlags.Resume = tc.resume

			err := stateMachine.validateInput()
			asserter.AssertErrContains(err, tc.errMsg)
		})
	}
}

// TestValidateUntilThru ensures that using invalid value for --thru
// or --until returns an error
func TestValidateUntilThru(t *testing.T) {
	testCases := []struct {
		name  string
		until string
		thru  string
	}{
		{"invalid_until_name", "fake step", ""},
		{"invalid_thru_name", "", "fake step"},
	}
	for _, tc := range testCases {
		t.Run("test "+tc.name, func(t *testing.T) {
			asserter := helper.Asserter{T: t}
			var stateMachine StateMachine
			stateMachine.commonFlags, stateMachine.stateMachineFlags = helper.InitCommonOpts()
			stateMachine.stateMachineFlags.Until = tc.until
			stateMachine.stateMachineFlags.Thru = tc.thru

			err := stateMachine.validateUntilThru()
			asserter.AssertErrContains(err, "not a valid state name")

		})
	}
}

// TestGenerateAptCmd unit tests the generateAptCmd function
func TestGenerateAptCmd(t *testing.T) {
	testCases := []struct {
		name        string
		targetDir   string
		packageList []string
		expected    string
	}{
		{"one_package", "chroot1", []string{"test"}, "chroot chroot1 apt install -y test"},
		{"many_packages", "chroot2", []string{"test1", "test2"}, "chroot chroot2 apt install -y test1 test2"},
	}
	for _, tc := range testCases {
		t.Run("test_generate_apt_cmd_"+tc.name, func(t *testing.T) {
			aptCmd := generateAptCmd(tc.targetDir, tc.packageList)
			if !strings.Contains(aptCmd.String(), tc.expected) {
				t.Errorf("Expected apt command \"%s\" but got \"%s\"", tc.expected, aptCmd.String())
			}
		})
	}
}

// TestCreatePPAInfo unit tests the createPPAInfo function
func TestCreatePPAInfo(t *testing.T) {
	testCases := []struct {
		name             string
		ppa              *PPAType
		series           string
		expectedName     string
		expectedContents string
	}{
		{
			"public_ppa",
			&PPAType{
				PPAName: "public/ppa",
			},
			"focal",
			"public-ubuntu-ppa-focal.list",
			"deb https://ppa.launchpadcontent.net/public/ppa/ubuntu focal main",
		},
		{
			"private_ppa",
			&PPAType{
				PPAName: "private/ppa",
				Auth:    "testuser:testpass",
			},
			"jammy",
			"private-ubuntu-ppa-jammy.list",
			"deb https://testuser:testpass@private-ppa.launchpadcontent.net/private/ppa/ubuntu jammy main",
		},
	}
	for _, tc := range testCases {
		t.Run("test_create_ppa_info_"+tc.name, func(t *testing.T) {
			fileName, fileContents := createPPAInfo(tc.ppa, tc.series)
			if fileName != tc.expectedName {
				t.Errorf("Expected PPA filename \"%s\" but got \"%s\"",
					tc.expectedName, fileName)
			}
			if fileContents != tc.expectedContents {
				t.Errorf("Expected PPA file contents \"%s\" but got \"%s\"",
					tc.expectedContents, fileContents)
			}
		})
	}
}
