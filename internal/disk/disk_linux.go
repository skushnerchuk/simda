//go:build linux

package disk

import (
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sys/unix"
)

const (
	AdfsSuperMagic      = 0xadf5
	AffsSuperMagic      = 0xADFF
	BdevfsMagic         = 0x62646576
	BefsSuperMagic      = 0x42465331
	BfsMagic            = 0x1BADFACE
	BinfmtfsMagic       = 0x42494e4d
	BtrfsSuperMagic     = 0x9123683E
	CgroupSuperMagic    = 0x27e0eb
	CifsMagicNumber     = 0xFF534D42
	CodaSuperMagic      = 0x73757245
	CohSuperMagic       = 0x012FF7B7
	CramfsMagic         = 0x28cd3d45
	DebugfsMagic        = 0x64626720
	DevfsSuperMagic     = 0x1373
	DevptsSuperMagic    = 0x1cd1
	EfivarfsMagic       = 0xde5e81e4
	EfsSuperMagic       = 0x00414A53
	ExtSuperMagic       = 0x137D
	Ext2OldSuperMagic   = 0xEF51
	Ext2SuperMagic      = 0xEF53
	Ext3SuperMagic      = 0xEF53
	Ext4SuperMagic      = 0xEF53
	FuseSuperMagic      = 0x65735546
	FutexfsSuperMagic   = 0xBAD1DEA
	HfsSuperMagic       = 0x4244
	HfsplusSuperMagic   = 0x482b
	HostfsSuperMagic    = 0x00c0ffee
	HpfsSuperMagic      = 0xF995E849
	HugetlbfsMagic      = 0x958458f6
	IsofsSuperMagic     = 0x9660
	Jffs2SuperMagic     = 0x72b6
	JfsSuperMagic       = 0x3153464a
	MinixSuperMagic     = 0x137F /* orig. minix */
	MinixSuperMagic2    = 0x138F /* 30 char minix */
	Minix2SuperMagic    = 0x2468 /* minix V2 */
	Minix2SuperMagic2   = 0x2478 /* minix V2, 30 char names */
	Minix3SuperMagic    = 0x4d5a /* minix V3 fs, 60 char names */
	MqueueMagic         = 0x19800202
	MsdosSuperMagic     = 0x4d44
	NcpSuperMagic       = 0x564c
	NfsSuperMagic       = 0x6969
	NilfsSuperMagic     = 0x3434
	NtfsSbMagic         = 0x5346544e
	Ocfs2SuperMagic     = 0x7461636f
	OpenpromSuperMagic  = 0x9fa1
	PipefsMagic         = 0x50495045
	ProcSuperMagic      = 0x9fa0
	PstorefsMagic       = 0x6165676C
	Qnx4SuperMagic      = 0x002f
	Qnx6SuperMagic      = 0x68191122
	RamfsMagic          = 0x858458f6
	ReiserfsSuperMagic  = 0x52654973
	RomfsMagic          = 0x7275
	SelinuxMagic        = 0xf97cff8c
	SmackMagic          = 0x43415d53
	SmbSuperMagic       = 0x517B
	SockfsMagic         = 0x534F434B
	SquashfsMagic       = 0x73717368
	SysfsMagic          = 0x62656572
	Sysv2SuperMagic     = 0x012FF7B6
	Sysv4SuperMagic     = 0x012FF7B5
	TmpfsMagic          = 0x01021994
	UdfSuperMagic       = 0x15013346
	UfsMagic            = 0x00011954
	UsbdeviceSuperMagic = 0x9fa2
	V9fsMagic           = 0x01021997
	VxfsSuperMagic      = 0xa501FCF5
	XenfsSuperMagic     = 0xabba1974
	XenixSuperMagic     = 0x012FF7B4
	XfsSuperMagic       = 0x58465342
	xiafsSuperMagic     = 0x012FD16D

	AfsSuperMagic            = 0x5346414F
	AufsSuperMagic           = 0x61756673
	AnonInodeFsSuperMagic    = 0x09041934
	BpfFsMagic               = 0xCAFE4A11
	CephSuperMagic           = 0x00C36400
	Cgroup2SuperMagic        = 0x63677270
	ConfigfsMagic            = 0x62656570
	EcryptfsSuperMagic       = 0xF15F
	F2fsSuperMagic           = 0xF2F52010
	FatSuperMagic            = 0x4006
	FhgfsSuperMagic          = 0x19830326
	FuseblkSuperMagic        = 0x65735546
	FusectlSuperMagic        = 0x65735543
	GfsSuperMagic            = 0x1161970
	GpfsSuperMagic           = 0x47504653
	MtdInodeFsSuperMagic     = 0x11307854
	InotifyfsSuperMagic      = 0x2BAD1DEA
	IsofsRWinSuperMagic      = 0x4004
	IsofsWinSuperMagic       = 0x4000
	JffsSuperMagic           = 0x07C0
	KafsSuperMagic           = 0x6B414653
	LustreSuperMagic         = 0x0BD00BD0
	NfsdSuperMagic           = 0x6E667364
	NsfsMagic                = 0x6E736673
	PanfsSuperMagic          = 0xAAD7AAEA
	RPCPipefsSuperMagic      = 0x67596969
	SecurityfsSuperMagic     = 0x73636673
	TracefsMagic             = 0x74726163
	UfsByteswappedSuperMagic = 0x54190100
	VmhgfsSuperMagic         = 0xBACBACBC
	VzfsSuperMagic           = 0x565A4653
	ZfsSuperMagic            = 0x2FC12FC1
)

var fsTypeMap = map[int64]string{
	AdfsSuperMagic:           "adfs",                /* 0xADF5 local */
	AffsSuperMagic:           "affs",                /* 0xADFF local */
	AfsSuperMagic:            "afs",                 /* 0x5346414F remote */
	AnonInodeFsSuperMagic:    "anon-inode FS",       /* 0x09041934 local */
	AufsSuperMagic:           "aufs",                /* 0x61756673 remote */
	BefsSuperMagic:           "befs",                /* 0x42465331 local */
	BdevfsMagic:              "bdevfs",              /* 0x62646576 local */
	BfsMagic:                 "bfs",                 /* 0x1BADFACE local */
	BinfmtfsMagic:            "binfmt_misc",         /* 0x42494E4D local */
	BpfFsMagic:               "bpf",                 /* 0xCAFE4A11 local */
	BtrfsSuperMagic:          "btrfs",               /* 0x9123683E local */
	CephSuperMagic:           "ceph",                /* 0x00C36400 remote */
	CgroupSuperMagic:         "cgroupfs",            /* 0x0027E0EB local */
	Cgroup2SuperMagic:        "cgroup2fs",           /* 0x63677270 local */
	CifsMagicNumber:          "cifs",                /* 0xFF534D42 remote */
	CodaSuperMagic:           "coda",                /* 0x73757245 remote */
	CohSuperMagic:            "coh",                 /* 0x012FF7B7 local */
	ConfigfsMagic:            "configfs",            /* 0x62656570 local */
	CramfsMagic:              "cramfs",              /* 0x28CD3D45 local */
	DebugfsMagic:             "debugfs",             /* 0x64626720 local */
	DevfsSuperMagic:          "devfs",               /* 0x1373 local */
	DevptsSuperMagic:         "devpts",              /* 0x1CD1 local */
	EcryptfsSuperMagic:       "ecryptfs",            /* 0xF15F local */
	EfivarfsMagic:            "efivarfs",            /* 0xDE5E81E4 local */
	EfsSuperMagic:            "efs",                 /* 0x00414A53 local */
	ExtSuperMagic:            "ext",                 /* 0x137D local */
	Ext2SuperMagic:           "ext2/ext3",           /* 0xEF53 local */
	Ext2OldSuperMagic:        "ext2",                /* 0xEF51 local */
	F2fsSuperMagic:           "f2fs",                /* 0xF2F52010 local */
	FatSuperMagic:            "fat",                 /* 0x4006 local */
	FhgfsSuperMagic:          "fhgfs",               /* 0x19830326 remote */
	FuseblkSuperMagic:        "fuseblk",             /* 0x65735546 remote */
	FusectlSuperMagic:        "fusectl",             /* 0x65735543 remote */
	FutexfsSuperMagic:        "futexfs",             /* 0x0BAD1DEA local */
	GfsSuperMagic:            "gfs/gfs2",            /* 0x1161970 remote */
	GpfsSuperMagic:           "gpfs",                /* 0x47504653 remote */
	HfsSuperMagic:            "hfs",                 /* 0x4244 local */
	HfsplusSuperMagic:        "hfsplus",             /* 0x482b local */
	HpfsSuperMagic:           "hpfs",                /* 0xF995E849 local */
	HugetlbfsMagic:           "hugetlbfs",           /* 0x958458F6 local */
	MtdInodeFsSuperMagic:     "inodefs",             /* 0x11307854 local */
	InotifyfsSuperMagic:      "inotifyfs",           /* 0x2BAD1DEA local */
	IsofsSuperMagic:          "isofs",               /* 0x9660 local */
	IsofsRWinSuperMagic:      "isofs",               /* 0x4004 local */
	IsofsWinSuperMagic:       "isofs",               /* 0x4000 local */
	JffsSuperMagic:           "jffs",                /* 0x07C0 local */
	Jffs2SuperMagic:          "jffs2",               /* 0x72B6 local */
	JfsSuperMagic:            "jfs",                 /* 0x3153464A local */
	KafsSuperMagic:           "k-afs",               /* 0x6B414653 remote */
	LustreSuperMagic:         "lustre",              /* 0x0BD00BD0 remote */
	MinixSuperMagic:          "minix",               /* 0x137F local */
	MinixSuperMagic2:         "minix (30 char.)",    /* 0x138F local */
	Minix2SuperMagic:         "minix v2",            /* 0x2468 local */
	Minix2SuperMagic2:        "minix v2 (30 char.)", /* 0x2478 local */
	Minix3SuperMagic:         "minix3",              /* 0x4D5A local */
	MqueueMagic:              "mqueue",              /* 0x19800202 local */
	MsdosSuperMagic:          "msdos",               /* 0x4D44 local */
	NcpSuperMagic:            "novell",              /* 0x564C remote */
	NfsSuperMagic:            "nfs",                 /* 0x6969 remote */
	NfsdSuperMagic:           "nfsd",                /* 0x6E667364 remote */
	NilfsSuperMagic:          "nilfs",               /* 0x3434 local */
	NsfsMagic:                "nsfs",                /* 0x6E736673 local */
	NtfsSbMagic:              "ntfs",                /* 0x5346544E local */
	OpenpromSuperMagic:       "openprom",            /* 0x9FA1 local */
	Ocfs2SuperMagic:          "ocfs2",               /* 0x7461636f remote */
	PanfsSuperMagic:          "panfs",               /* 0xAAD7AAEA remote */
	PipefsMagic:              "pipefs",              /* 0x50495045 remote */
	ProcSuperMagic:           "proc",                /* 0x9FA0 local */
	PstorefsMagic:            "pstorefs",            /* 0x6165676C local */
	Qnx4SuperMagic:           "qnx4",                /* 0x002F local */
	Qnx6SuperMagic:           "qnx6",                /* 0x68191122 local */
	RamfsMagic:               "ramfs",               /* 0x858458F6 local */
	ReiserfsSuperMagic:       "reiserfs",            /* 0x52654973 local */
	RomfsMagic:               "romfs",               /* 0x7275 local */
	RPCPipefsSuperMagic:      "rpc_pipefs",          /* 0x67596969 local */
	SecurityfsSuperMagic:     "securityfs",          /* 0x73636673 local */
	SelinuxMagic:             "selinux",             /* 0xF97CFF8C local */
	SmbSuperMagic:            "smb",                 /* 0x517B remote */
	SockfsMagic:              "sockfs",              /* 0x534F434B local */
	SquashfsMagic:            "squashfs",            /* 0x73717368 local */
	SysfsMagic:               "sysfs",               /* 0x62656572 local */
	Sysv2SuperMagic:          "sysv2",               /* 0x012FF7B6 local */
	Sysv4SuperMagic:          "sysv4",               /* 0x012FF7B5 local */
	TmpfsMagic:               "tmpfs",               /* 0x01021994 local */
	TracefsMagic:             "tracefs",             /* 0x74726163 local */
	UdfSuperMagic:            "udf",                 /* 0x15013346 local */
	UfsMagic:                 "ufs",                 /* 0x00011954 local */
	UfsByteswappedSuperMagic: "ufs",                 /* 0x54190100 local */
	UsbdeviceSuperMagic:      "usbdevfs",            /* 0x9FA2 local */
	V9fsMagic:                "v9fs",                /* 0x01021997 local */
	VmhgfsSuperMagic:         "vmhgfs",              /* 0xBACBACBC remote */
	VxfsSuperMagic:           "vxfs",                /* 0xA501FCF5 local */
	VzfsSuperMagic:           "vzfs",                /* 0x565A4653 local */
	XenfsSuperMagic:          "xenfs",               /* 0xABBA1974 local */
	XenixSuperMagic:          "xenix",               /* 0x012FF7B4 local */
	XfsSuperMagic:            "xfs",                 /* 0x58465342 local */
	xiafsSuperMagic:          "xia",                 /* 0x012FD16D local */
	ZfsSuperMagic:            "zfs",                 /* 0x2FC12FC1 local */
}

func GetFsType(stat unix.Statfs_t) string {
	ret, ok := fsTypeMap[stat.Type]
	if !ok {
		return ""
	}
	return ret
}

func UnescapeFstab(path string) string {
	escaped, err := strconv.Unquote(`"` + path + `"`)
	if err != nil {
		return path
	}
	return escaped
}

func GetDevices(sysPath string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(sysPath, "block"))
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	return files, nil
}
