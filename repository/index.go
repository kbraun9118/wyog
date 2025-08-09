package repository

import "time"

type IndexHeader struct {
	Signature [4]byte
	Version   int32
	Count     int32
}

type IndexBinaryEntry struct {
	CtimeSec  uint32
	CtimeNSec uint32
	MtimeSec  uint32
	MtimeNSec uint32
	Dev       uint32
	Ino       uint32
	Unused    uint16
	Mode      uint16
	Uid       uint32
	Gid       uint32
	Fsize     uint32
	Sha       [20]byte
	Flags     uint16
}

type IndexEntry struct {
	Ctime       time.Time
	Mtime       time.Time
	Dev         int
	Ino         int
	ModeType    int
	ModePerms   int
	Uid         int
	Gid         int
	Fsize       int
	Sha         string
	AssumeValid bool
	Stage       int
	Name        string
}

type Index struct {
	Version int
	Entries []IndexEntry
}

func NewIndexV2(entries []IndexEntry) Index {
	if entries == nil {
		entries = make([]IndexEntry, 0)
	}

	return Index{
		Version: 2,
		Entries: entries,
	}
}
