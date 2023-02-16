package lru

type Node struct {
	key, value string
	prev, next *Node
}

type cachrLru struct {
	size       int
	capacity   int
	list       map[string]*Node
	head, tail *Node
}

// InitLru 初始化新的lru,返回链表需要常态储存
func InitLru(capacity int) *cachrLru {
	head := initLruCache("", "")
	tail := initLruCache("", "")
	head.next = tail
	tail.prev = head
	return &cachrLru{
		size:     0,
		capacity: capacity,
		list:     make(map[string]*Node),
		head:     head,
		tail:     tail,
	}
}

func (this *cachrLru) Get(key string) string {
	if _, ok := this.list[key]; !ok {
		return "Not Found"
	}
	node := this.list[key]
	this.moveToHead(node)
	return node.value
}

func (this *cachrLru) Put(key string, value string) {
	if _, ok := this.list[key]; ok {
		//已有.更新
		node := this.list[key]
		node.value = value
		this.moveToHead(node)
	} else {
		//没有.插入
		node := initLruCache(key, value)
		this.addHead(node)
		if this.size > this.capacity {
			this.delLast()
		}
	}
}

func (this *cachrLru) DelNodeByKey(key string) {
	node := this.list[key]
	if node != nil {
		this.delNode(node)
	}
}

func (this *cachrLru) moveToHead(node *Node) {
	this.delNode(node)
	this.addHead(node)
}

func (this *cachrLru) addHead(node *Node) {
	node.prev = this.head
	node.next = this.head.next
	node.next.prev = node
	this.head.next = node
	this.list[node.key] = node
	this.size++
}

func (this *cachrLru) delNode(node *Node) {
	node.prev.next = node.next
	node.next.prev = node.prev
	delete(this.list, node.key)
	this.size--
}

func (this *cachrLru) delLast() {
	node := this.tail.prev
	this.delNode(node)
}

func initLruCache(key, value string) *Node {
	return &Node{
		key:   key,
		value: value,
	}
}
