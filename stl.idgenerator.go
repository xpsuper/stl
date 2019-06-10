package stl

import (
	"sync"
	"time"
)

const (
	kBasicTimestamp = 1483200000000 // 2017/1/1 00:00:00
	kWorkerIdBits   = 10            // Num of WorkerId Bits
	kSequenceBits   = 12            // Num of Sequence Bits

	kWorkerIdShift  = 12
	kTimeStampShift = 22

	kSequenceMask   = 0xfff // equal as getSequenceMask()
	kMaxWorker      = 0x3ff // equal as getMaxWorkerId()
)

type XPIdGeneratorImpl struct {
	workerId      int64
	lastTimeStamp int64
	sequence      int64
	maxWorkerId   int64
	lock          *sync.Mutex
}

func NewIdGenerator(workerId int64) (idGenerator *XPIdGeneratorImpl, err error) {
	idGenerator = new(XPIdGeneratorImpl)

	idGenerator.maxWorkerId = getMaxWorkerId()

	if workerId > idGenerator.maxWorkerId || workerId < 0 {
		return nil, NewErrors("worker not fit")
	}
	idGenerator.workerId = workerId
	idGenerator.lastTimeStamp = -1
	idGenerator.sequence = 0
	idGenerator.lock = new(sync.Mutex)
	return idGenerator, nil
}

func (idGenerator *XPIdGeneratorImpl) NextId() (id int64, err error) {
	idGenerator.lock.Lock()
	defer idGenerator.lock.Unlock()
	id = idGenerator.timeGen()
	if id == idGenerator.lastTimeStamp {
		idGenerator.sequence = (idGenerator.sequence + 1) & kSequenceMask
		if idGenerator.sequence == 0 {
			id = idGenerator.timeReGen(id)
		}
	} else {
		idGenerator.sequence = 0
	}

	if id < idGenerator.lastTimeStamp {
		return 0, NewErrors("Clock moved backwards, Refuse gen id")
	}
	idGenerator.lastTimeStamp = id
	id = (id-kBasicTimestamp)<<kTimeStampShift | idGenerator.workerId<<kWorkerIdShift | idGenerator.sequence
	return id, nil
}

func (idGenerator *XPIdGeneratorImpl) ParseId(id int64) (t time.Time, ts int64, workerId int64, seq int64) {
	seq = id & kSequenceMask
	workerId = (id >> kWorkerIdShift) & kMaxWorker
	ts = (id >> kTimeStampShift) + kBasicTimestamp
	t = time.Unix(ts/1000, (ts%1000)*1000000)
	return
}

func getMaxWorkerId() int64 {
	return -1 ^ -1<<kWorkerIdBits
}

func getSequenceMask() int64 {
	return -1 ^ -1<<kSequenceBits
}

func (idGenerator *XPIdGeneratorImpl) timeGen() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}

func (idGenerator *XPIdGeneratorImpl) timeReGen(last int64) int64 {
	ts := time.Now().UnixNano()
	for {
		if ts < last {
			ts = idGenerator.timeGen()
		} else {
			break
		}
	}
	return ts
}
