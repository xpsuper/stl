package stl

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希函数
type Hash func(data []byte) uint32

// ConsistencyHash 一致性Hash
// 注意，非线程安全，业务需要自行加锁
type ConsistencyHash struct {
	hash Hash
	// 每个真实节点的虚拟节点数量
	replicas int
	// 哈希环，按照节点哈希值排序
	ring []int
	// 节点哈希值到真实节点字符串，哈希映射的逆过程
	nodes map[int]string
}

func NewConsistencyHash(replicas int, fn Hash) *ConsistencyHash {
	r := &ConsistencyHash{
		replicas: replicas,
		hash:     fn,
		nodes:    make(map[int]string),
	}
	if r.hash == nil {
		r.hash = crc32.ChecksumIEEE
	}
	return r
}

// Empty 哈希环上是否有节点
func (c *ConsistencyHash) Empty() bool {
	return len(c.ring) == 0
}

// Add 添加新节点到哈希环
// 注意，如果加入的节点已经存在，会导致哈希环上面重复，如果不确定是否存在请使用Reset
func (c *ConsistencyHash) Add(nodes ...string) {
	for _, node := range nodes {
		// 每个节点创建多个虚拟节点
		for i := 0; i < c.replicas; i++ {
			// 每个虚拟节点计算哈希值
			hash := int(c.hash([]byte(strconv.Itoa(i) + node)))
			// 加入哈希环
			c.ring = append(c.ring, hash)
			// 哈希值到真实节点字符串映射
			c.nodes[hash] = node
		}
	}
	// 哈希环排序
	sort.Ints(c.ring)
}

// Reset 先清空哈希环再设置
func (c *ConsistencyHash) Reset(nodes ...string) {
	// 先清空
	c.ring = nil
	c.nodes = map[int]string{}
	// 再重置
	c.Add(nodes...)
}

// Get 获取Key对应的节点
func (c *ConsistencyHash) Get(key string) string {
	// 如果哈希环位空，则直接返回
	if c.Empty() {
		return ""
	}

	// 计算Key哈希值
	hash := int(c.hash([]byte(key)))

	// 二分查找第一个大于等于Key哈希值的节点
	idx := sort.Search(len(c.ring), func(i int) bool { return c.ring[i] >= hash })

	// 这里是特殊情况，也就是数组没有大于等于Key哈希值的节点
	// 但是逻辑上这是一个环，因此第一个节点就是目标节点
	if idx == len(c.ring) {
		idx = 0
	}

	// 返回哈希值对应的真实节点字符串
	return c.nodes[c.ring[idx]]
}
