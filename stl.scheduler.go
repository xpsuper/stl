package stl

/*********调用示例********
func task() {
    fmt.Println("测试运行....")
}

func taskWithParams(a string, b string) {
    fmt.Println(a, b)
}

func main() {
    worker := NewScheduler()

    worker.Every(1).Second().Do(taskWithParams, 1, "hello")

    worker.Every(1).Day().At("18:56").Do(task)

    //调度器启动
    worker.Start()
}
 ************************/

import (
	"sort"
	"time"
	"log"
	"math"
	"strconv"
	"strings"
)

type Job interface {
	Run()
}

type Schedule interface {
	Next(time.Time) time.Time
}

type Entry struct {
	Name string
	Schedule Schedule
	Next time.Time
	Prev time.Time
	Job Job
}

type XPSchedulerImpl struct {
	entries  entries
	stop     chan struct{}
	add      chan *Entry
	remove   chan string
	snapshot chan entries
	running  bool
}

type bounds struct {
	min, max uint
	names    map[string]uint
}

type entries []*Entry
type byTime []*Entry

var (
	seconds = bounds{0, 59, nil}
	minutes = bounds{0, 59, nil}
	hours   = bounds{0, 23, nil}
	dom     = bounds{1, 31, nil}
	months  = bounds{1, 12, map[string]uint{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	dow = bounds{0, 6, map[string]uint{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)

const (
	starBit = 1 << 63
)

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

func (entrySlice entries) pos(name string) int {
	for p, e := range entrySlice {
		if e.Name == name {
			return p
		}
	}
	return -1
}

type FuncJob func()
func (f FuncJob) Run() { f() }

/**************** SpecSchedule ****************/
type SpecSchedule struct {
	Second, Minute, Hour, Dom, Month, Dow uint64
}

func dayMatches(s *SpecSchedule, t time.Time) bool {
	var (
		domMatch bool = 1<<uint(t.Day())&s.Dom > 0
		dowMatch bool = 1<<uint(t.Weekday())&s.Dow > 0
	)

	if s.Dom&starBit > 0 || s.Dow&starBit > 0 {
		return domMatch && dowMatch
	}
	return domMatch || dowMatch
}

func (s *SpecSchedule) Next(t time.Time) time.Time {
	// Start at the earliest possible time (the upcoming second).
	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)

	// This flag indicates whether a field has been incremented.
	added := false

	// If no time is found within five years, return zero.
	yearLimit := t.Year() + 5

WRAP:
	if t.Year() > yearLimit {
		return time.Time{}
	}

	// Find the first applicable month.
	// If it's this month, then do nothing.
	for 1<<uint(t.Month())&s.Month == 0 {
		// If we have to add a month, reset the other parts to 0.
		if !added {
			added = true
			// Otherwise, set the date at the beginning (since the current time is irrelevant).
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 1, 0)

		// Wrapped around.
		if t.Month() == time.January {
			goto WRAP
		}
	}

	// Now get a day in that month.
	for !dayMatches(s, t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 0, 1)

		if t.Day() == 1 {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&s.Hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
		}
		t = t.Add(1 * time.Hour)

		if t.Hour() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
		}
		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Second())&s.Second == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
		}
		t = t.Add(1 * time.Second)

		if t.Second() == 0 {
			goto WRAP
		}
	}

	return t
}
/**************** SpecSchedule ****************/

/**************** ConstantDelaySchedule ****************/
type ConstantDelaySchedule struct {
	Delay time.Duration
}

func Every(duration time.Duration) ConstantDelaySchedule {
	if duration < time.Second {
		panic("cron/constantdelay: delays of less than a second are not supported: " +
			duration.String())
	}
	return ConstantDelaySchedule{
		Delay: duration - time.Duration(duration.Nanoseconds())%time.Second,
	}
}

func (schedule ConstantDelaySchedule) Next(t time.Time) time.Time {
	return t.Add(schedule.Delay - time.Duration(t.Nanosecond())*time.Nanosecond)
}
/**************** ConstantDelaySchedule ****************/

/**************** cron表达式解析 ****************/
func getField(field string, r bounds) uint64 {
	// list = range {"," range}
	var bits uint64
	ranges := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	for _, expr := range ranges {
		bits |= getRange(expr, r)
	}
	return bits
}

func getRange(expr string, r bounds) uint64 {

	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep[0], "-")
		singleDigit      = len(lowAndHigh) == 1
	)

	var extra_star uint64
	if lowAndHigh[0] == "*" || lowAndHigh[0] == "?" {
		start = r.min
		end = r.max
		extra_star = starBit
	} else {
		start = parseIntOrName(lowAndHigh[0], r.names)
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			end = parseIntOrName(lowAndHigh[1], r.names)
		default:
			log.Panicf("Too many hyphens: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step = mustParseInt(rangeAndStep[1])

		// Special handling: "N/step" means "N-max/step".
		if singleDigit {
			end = r.max
		}
	default:
		log.Panicf("Too many slashes: %s", expr)
	}

	if start < r.min {
		log.Panicf("Beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		log.Panicf("End of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		log.Panicf("Beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}

	return getBits(start, end, step) | extra_star
}

func parseIntOrName(expr string, names map[string]uint) uint {
	if names != nil {
		if namedInt, ok := names[strings.ToLower(expr)]; ok {
			return namedInt
		}
	}
	return mustParseInt(expr)
}

func mustParseInt(expr string) uint {
	num, err := strconv.Atoi(expr)
	if err != nil {
		log.Panicf("Failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		log.Panicf("Negative number (%d) not allowed: %s", num, expr)
	}

	return uint(num)
}

func getBits(min, max, step uint) uint64 {
	var bits uint64

	// If step is 1, use shifts.
	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	// Else, use a simple loop.
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

func all(r bounds) uint64 {
	return getBits(r.min, r.max, 1) | starBit
}

func parseDescriptor(spec string) Schedule {
	switch spec {
	case "@yearly", "@annually":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  1 << months.min,
			Dow:    all(dow),
		}

	case "@monthly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  all(months),
			Dow:    all(dow),
		}

	case "@weekly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    1 << dow.min,
		}

	case "@daily", "@midnight":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}

	case "@hourly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   all(hours),
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}
	}

	const every = "@every "
	if strings.HasPrefix(spec, every) {
		duration, err := time.ParseDuration(spec[len(every):])
		if err != nil {
			log.Panicf("Failed to parse duration %s: %s", spec, err)
		}
		return Every(duration)
	}

	log.Panicf("Unrecognized descriptor: %s", spec)
	return nil
}

func parseCron(spec string) Schedule {
	if spec[0] == '@' {
		return parseDescriptor(spec)
	}

	fields := strings.Fields(spec)
	if len(fields) != 5 && len(fields) != 6 {
		log.Panicf("Expected 5 or 6 fields, found %d: %s", len(fields), spec)
	}

	// If a sixth field is not provided (DayOfWeek), then it is equivalent to star.
	if len(fields) == 5 {
		fields = append(fields, "*")
	}

	schedule := &SpecSchedule{
		Second: getField(fields[0], seconds),
		Minute: getField(fields[1], minutes),
		Hour:   getField(fields[2], hours),
		Dom:    getField(fields[3], dom),
		Month:  getField(fields[4], months),
		Dow:    getField(fields[5], dow),
	}

	return schedule
}
/**************** cron表达式解析 ****************/

func NewScheduler() *XPSchedulerImpl {
	return &XPSchedulerImpl{
		entries:  nil,
		add:      make(chan *Entry),
		remove:   make(chan string),
		stop:     make(chan struct{}),
		snapshot: make(chan entries),
		running:  false,
	}
}

func (instance *XPSchedulerImpl) AddFunc(spec string, cmd func(), name string) {
	instance.AddJob(spec, FuncJob(cmd), name)
}

func (instance *XPSchedulerImpl) AddJob(spec string, cmd Job, name string) {
	instance.Schedule(parseCron(spec), cmd, name)
}

func (instance *XPSchedulerImpl) RemoveJob(name string) {
	if !instance.running {
		i := instance.entries.pos(name)

		if i == -1 {
			return
		}

		instance.entries = instance.entries[:i+copy(instance.entries[i:], instance.entries[i+1:])]
		return
	}

	instance.remove <- name
}

func (instance *XPSchedulerImpl) Schedule(schedule Schedule, cmd Job, name string) {
	entry := &Entry{
		Schedule: schedule,
		Job:      cmd,
		Name:     name,
	}

	if !instance.running {
		i := instance.entries.pos(entry.Name)
		if i != -1 {
			return
		}
		instance.entries = append(instance.entries, entry)
		return
	}

	instance.add <- entry
}

func (instance *XPSchedulerImpl) Entries() []*Entry {
	if instance.running {
		instance.snapshot <- nil
		x := <-instance.snapshot
		return x
	}
	return instance.entrySnapshot()
}

func (instance *XPSchedulerImpl) Start() {
	instance.running = true
	go instance.run()
}

func (instance *XPSchedulerImpl) Stop() {
	instance.stop <- struct{}{}
	instance.running = false
}

func (instance *XPSchedulerImpl) StartAsService() {
	instance.running = true
	instance.run()
}

func (instance *XPSchedulerImpl) run() {
	// Figure out the next activation times for each entry.
	now := time.Now().Local()
	for _, entry := range instance.entries {
		entry.Next = entry.Schedule.Next(now)
	}

	for {
		// Determine the next entry to run.
		sort.Sort(byTime(instance.entries))

		var effective time.Time
		if len(instance.entries) == 0 || instance.entries[0].Next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			effective = now.AddDate(10, 0, 0)
		} else {
			effective = instance.entries[0].Next
		}

		select {
		case now = <-time.After(effective.Sub(now)):
			// Run every entry whose next time was this effective time.
			for _, e := range instance.entries {
				if e.Next != effective {
					break
				}
				go e.Job.Run()
				e.Prev = e.Next
				e.Next = e.Schedule.Next(effective)
			}
			continue

		case newEntry := <-instance.add:
			i := instance.entries.pos(newEntry.Name)
			if i != -1 {
				break
			}
			instance.entries = append(instance.entries, newEntry)
			newEntry.Next = newEntry.Schedule.Next(time.Now().Local())

		case name := <-instance.remove:
			i := instance.entries.pos(name)

			if i == -1 {
				break
			}

			instance.entries = instance.entries[:i+copy(instance.entries[i:], instance.entries[i+1:])]

		case <-instance.snapshot:
			instance.snapshot <- instance.entrySnapshot()

		case <-instance.stop:
			return
		}

		// 'now' should be updated after newEntry and snapshot cases.
		now = time.Now().Local()
	}
}

func (instance *XPSchedulerImpl) entrySnapshot() []*Entry {
	entries := []*Entry{}
	for _, e := range instance.entries {
		entries = append(entries, &Entry{
			Schedule: e.Schedule,
			Next:     e.Next,
			Prev:     e.Prev,
			Job:      e.Job,
		})
	}
	return entries
}
