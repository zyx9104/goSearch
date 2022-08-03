package tree

type Queue []*AcNode

func (q Queue) Front() *AcNode {
	return q[0]
}

func (q *Queue) Push(node *AcNode) {
	*q = append(*q, node)
}

func (q *Queue) Pop() {
	*q = (*q)[1:]
}

func (q Queue) Empty() bool {
	return len(q) == 0
}

const charSet = 256

// AcNode AC自动机节点结构定义
type AcNode struct {
	cnt  int              // 结束模式串个数
	fail *AcNode          // fail指针节点
	son  [charSet]*AcNode // 子节点
}

func NewAC() *AcNode {
	return &AcNode{}
}

func (ac *AcNode) insert(s string) {

	bytes := []byte(s)
	u := ac
	for _, c := range bytes {
		if u.son[c] == nil {
			u.son[c] = &AcNode{}
		}
		u = u.son[c]
	}
	if u != ac {
		u.cnt++
	}

}

func (ac *AcNode) Build(strs []string) {
	for _, str := range strs {
		ac.insert(str)
	}
	q := &Queue{}
	for i := 0; i < charSet; i++ {
		if ac.son[i] != nil {
			ac.son[i].fail = ac
			q.Push(ac.son[i])
		} else {
			ac.son[i] = ac
		}
	}
	for !q.Empty() {
		u := q.Front()
		q.Pop()
		for i := 0; i < charSet; i++ {
			if u.son[i] != nil {
				u.son[i].fail = u.fail.son[i]
				q.Push(u.son[i])
			} else {
				u.son[i] = u.fail.son[i]
			}
		}
	}
}

func (ac *AcNode) Find(s string) bool {

	u := ac
	str := []byte(s)
	for _, c := range str {
		u = u.son[c]
		for v := u; v != nil; v = v.fail {
			if v.cnt != 0 {
				return true
			}
		}

	}
	return false
}
