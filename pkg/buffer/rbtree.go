package buffer

// Color represents the color of a node in the red-black tree
type Color bool

const (
	Red   Color = true
	Black Color = false
)

// Node represents a node in the red-black tree
type Node struct {
	Key    int         // Position in the text
	Value  interface{} // Data stored at this position
	Color  Color
	Left   *Node
	Right  *Node
	Parent *Node
}

// RBTree represents a red-black tree
type RBTree struct {
	Root *Node
	Nil  *Node // sentinel nil node
}

// NewRBTree creates a new red-black tree
func NewRBTree() *RBTree {
	nil := &Node{Color: Black}
	return &RBTree{
		Root: nil,
		Nil:  nil,
	}
}

// Search finds a node with the given key in the tree
func (t *RBTree) Search(key int) *Node {
	return t.searchTreeHelper(t.Root, key)
}

// searchTreeHelper is a helper function for Search
func (t *RBTree) searchTreeHelper(node *Node, key int) *Node {
	if node == t.Nil {
		return nil
	}

	if key == node.Key {
		return node
	}

	if key < node.Key {
		return t.searchTreeHelper(node.Left, key)
	}

	return t.searchTreeHelper(node.Right, key)
}

// Insert adds a new node with the given key and value to the tree
func (t *RBTree) Insert(key int, value interface{}) {
	// Create new node
	newNode := &Node{
		Key:    key,
		Value:  value,
		Color:  Red,
		Left:   t.Nil,
		Right:  t.Nil,
		Parent: t.Nil,
	}

	var y *Node = t.Nil
	var x *Node = t.Root

	// Find position for new node
	for x != t.Nil {
		y = x
		if newNode.Key < x.Key {
			x = x.Left
		} else {
			x = x.Right
		}
	}

	// Set parent of new node
	newNode.Parent = y
	if y == t.Nil {
		// Tree was empty
		t.Root = newNode
	} else if newNode.Key < y.Key {
		y.Left = newNode
	} else {
		y.Right = newNode
	}

	// If new node is root, color it black and return
	if newNode.Parent == t.Nil {
		newNode.Color = Black
		return
	}

	// If grandparent is nil, return
	if newNode.Parent.Parent == t.Nil {
		return
	}

	// Fix red-black tree properties
	t.fixInsert(newNode)
}

// LeftRotate performs a left rotation on the given node
func (t *RBTree) leftRotate(x *Node) {
	y := x.Right
	x.Right = y.Left
	if y.Left != t.Nil {
		y.Left.Parent = x
	}
	y.Parent = x.Parent
	if x.Parent == t.Nil {
		t.Root = y
	} else if x == x.Parent.Left {
		x.Parent.Left = y
	} else {
		x.Parent.Right = y
	}
	y.Left = x
	x.Parent = y
}

// RightRotate performs a right rotation on the given node
func (t *RBTree) rightRotate(x *Node) {
	y := x.Left
	x.Left = y.Right
	if y.Right != t.Nil {
		y.Right.Parent = x
	}
	y.Parent = x.Parent
	if x.Parent == t.Nil {
		t.Root = y
	} else if x == x.Parent.Right {
		x.Parent.Right = y
	} else {
		x.Parent.Left = y
	}
	y.Right = x
	x.Parent = y
}

// fixInsert fixes the red-black tree properties after insertion
func (t *RBTree) fixInsert(k *Node) {
	var u *Node
	for k.Parent.Color == Red {
		if k.Parent == k.Parent.Parent.Right {
			u = k.Parent.Parent.Left
			if u.Color == Red {
				u.Color = Black
				k.Parent.Color = Black
				k.Parent.Parent.Color = Red
				k = k.Parent.Parent
			} else {
				if k == k.Parent.Left {
					k = k.Parent
					t.rightRotate(k)
				}
				k.Parent.Color = Black
				k.Parent.Parent.Color = Red
				t.leftRotate(k.Parent.Parent)
			}
		} else {
			u = k.Parent.Parent.Right
			if u.Color == Red {
				u.Color = Black
				k.Parent.Color = Black
				k.Parent.Parent.Color = Red
				k = k.Parent.Parent
			} else {
				if k == k.Parent.Right {
					k = k.Parent
					t.leftRotate(k)
				}
				k.Parent.Color = Black
				k.Parent.Parent.Color = Red
				t.rightRotate(k.Parent.Parent)
			}
		}
		if k == t.Root {
			break
		}
	}
	t.Root.Color = Black
}

// Delete removes a node with the given key from the tree
func (t *RBTree) Delete(key int) {
	t.deleteNodeHelper(t.Root, key)
}

// deleteNodeHelper is a helper function for Delete
func (t *RBTree) deleteNodeHelper(node *Node, key int) {
	z := t.Nil
	var x, y *Node

	// Find the node to delete
	for node != t.Nil {
		if node.Key == key {
			z = node
			break
		}

		if node.Key < key {
			node = node.Right
		} else {
			node = node.Left
		}
	}

	if z == t.Nil {
		return
	}

	y = z
	originalColor := y.Color

	if z.Left == t.Nil {
		x = z.Right
		t.transplant(z, z.Right)
	} else if z.Right == t.Nil {
		x = z.Left
		t.transplant(z, z.Left)
	} else {
		y = t.minimum(z.Right)
		originalColor = y.Color
		x = y.Right

		if y.Parent == z {
			x.Parent = y
		} else {
			t.transplant(y, y.Right)
			y.Right = z.Right
			y.Right.Parent = y
		}

		t.transplant(z, y)
		y.Left = z.Left
		y.Left.Parent = y
		y.Color = z.Color
	}

	if originalColor == Black {
		t.fixDelete(x)
	}
}

// transplant replaces one subtree with another
func (t *RBTree) transplant(u, v *Node) {
	if u.Parent == t.Nil {
		t.Root = v
	} else if u == u.Parent.Left {
		u.Parent.Left = v
	} else {
		u.Parent.Right = v
	}
	v.Parent = u.Parent
}

// minimum finds the node with the minimum key in the subtree rooted at node
func (t *RBTree) minimum(node *Node) *Node {
	for node.Left != t.Nil {
		node = node.Left
	}
	return node
}

// fixDelete fixes the red-black tree properties after deletion
func (t *RBTree) fixDelete(x *Node) {
	var s *Node
	for x != t.Root && x.Color == Black {
		if x == x.Parent.Left {
			s = x.Parent.Right
			if s.Color == Red {
				s.Color = Black
				x.Parent.Color = Red
				t.leftRotate(x.Parent)
				s = x.Parent.Right
			}

			if s.Left.Color == Black && s.Right.Color == Black {
				s.Color = Red
				x = x.Parent
			} else {
				if s.Right.Color == Black {
					s.Left.Color = Black
					s.Color = Red
					t.rightRotate(s)
					s = x.Parent.Right
				}

				s.Color = x.Parent.Color
				x.Parent.Color = Black
				s.Right.Color = Black
				t.leftRotate(x.Parent)
				x = t.Root
			}
		} else {
			s = x.Parent.Left
			if s.Color == Red {
				s.Color = Black
				x.Parent.Color = Red
				t.rightRotate(x.Parent)
				s = x.Parent.Left
			}

			if s.Right.Color == Black && s.Left.Color == Black {
				s.Color = Red
				x = x.Parent
			} else {
				if s.Left.Color == Black {
					s.Right.Color = Black
					s.Color = Red
					t.leftRotate(s)
					s = x.Parent.Left
				}

				s.Color = x.Parent.Color
				x.Parent.Color = Black
				s.Left.Color = Black
				t.rightRotate(x.Parent)
				x = t.Root
			}
		}
	}
	x.Color = Black
}

// InOrderTraversal performs an in-order traversal of the tree and applies the given function to each node
func (t *RBTree) InOrderTraversal(fn func(key int, value interface{})) {
	t.inOrderHelper(t.Root, fn)
}

// inOrderHelper is a helper function for InOrderTraversal
func (t *RBTree) inOrderHelper(node *Node, fn func(key int, value interface{})) {
	if node != t.Nil {
		t.inOrderHelper(node.Left, fn)
		fn(node.Key, node.Value)
		t.inOrderHelper(node.Right, fn)
	}
}

// Update updates the value of a node with the given key
func (t *RBTree) Update(key int, value interface{}) bool {
	node := t.Search(key)
	if node == nil {
		return false
	}
	node.Value = value
	return true
}
