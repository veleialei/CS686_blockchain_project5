package p1

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Flag_value struct {
	encoded_prefix []uint8
	value          string
}

type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	db    map[string]Node
	plain map[string]string
	root  string
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}
	address := string(node.flag_value.encoded_prefix) + node.flag_value.value
	str = address + str
	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
	mpt.plain = make(map[string]string)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Get_root() string {
	fmt.Println("What????????????????????????????????????????", mpt.root)
	return mpt.root
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	if len(mpt.db) == 0 {
		return "empty"
	}
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}

//=================================== helper functions ===================================
// change string to hex array for encode and create path
func string2hexarray(key string) []uint8 {
	ascii := []uint8{}
	for _, r := range key {
		ascii = append(ascii, uint8(r))
	}
	ret := []uint8{}
	for _, r := range ascii {
		p := r / 16
		q := r % 16
		ret = append(ret, p)
		ret = append(ret, q)
	}
	return ret
}

// judge if is a leaf node
func is_leaf(prefix []uint8) bool {
	if prefix == nil {
		return false
	}
	return prefix[0]/16 > 1
}

// get the eq length of two array
func eq_len(a, b []uint8) int {
	size := 0
	if (a == nil) != (b == nil) {
		return 0
	}
	if len(a) > len(b) {
		return eq_len(b, a)
	}
	for i := range a {
		if a[i] != b[i] {
			break
		}
		size++
	}
	return size
}

// return the equal part, and different parts of two array
func path_compare(path []uint8, prefix []uint8) ([]uint8, []uint8, []uint8) {
	eq_len := eq_len(path, prefix)
	eq_part := path[:eq_len]
	re_prefix := prefix[eq_len:]
	re_path := path[eq_len:]
	return eq_part, re_prefix, re_path
}

// calculate the size of branch
func (mpt *MerklePatriciaTrie) branchSize(key string) int {
	node := mpt.db[key]
	size := 0
	for i := 0; i < 17; i++ {
		if node.branch_value[i] != "" {
			size++
		}
	}
	return size
}

// helper function to print all data in mpt.db
func (mpt *MerklePatriciaTrie) PrintDB() {
	if mpt.root == "" {
		fmt.Println("DB is empty")
		return
	}
	fmt.Println("size of db: ", len(mpt.db))
	for k, v := range mpt.db {
		fmt.Println("Key is : " + k)
		node_type := v.node_type
		if node_type == 0 {
			fmt.Println("Null Node")
		} else if node_type == 1 {
			fmt.Println("Branch Node")
			branch_value := v.branch_value
			fmt.Println(branch_value)
		} else {
			prefix := v.flag_value.encoded_prefix
			fmt.Println(prefix)
			if prefix[0]/16 > 1 {
				fmt.Println("Leaf Node")
			} else {
				fmt.Println("Ext Node")
			}
			fmt.Println(v.flag_value.value)
		}
		fmt.Println("========================================================")
	}
}

// create a new branchnode
func (mpt *MerklePatriciaTrie) newBranchNode(branch_value [17]string) string {
	node := Node{node_type: 1, branch_value: branch_value}
	hash := node.hash_node()
	mpt.db[hash] = node
	return hash
}

// create a new leaf node
func (mpt *MerklePatriciaTrie) newLeafExtNode(path []uint8, new_value string, leaf bool) string {
	if leaf {
		path = append(path, 16)
	}
	encoded := compact_encode(path)
	flag := Flag_value{encoded_prefix: encoded, value: new_value}
	node := Node{node_type: 2, flag_value: flag}
	hash := node.hash_node()
	mpt.db[hash] = node
	return hash
}

// update hash in db
func (mpt *MerklePatriciaTrie) UpdateHash(new_hash string, cur_hash string, prev_hash string) {
	if new_hash == cur_hash {
		return
	}
	if prev_hash == "" && mpt.root == cur_hash {
		mpt.root = new_hash

	} else {
		prev_node := mpt.db[prev_hash]
		switch prev_node.node_type {
		case 1: // Branch
			for i := 0; i < 16; i++ {
				if prev_node.branch_value[i] == cur_hash {
					prev_node.branch_value[i] = new_hash
					break
				}
			}
		case 2: // Ext
			prev_node.flag_value.value = new_hash
		}
		mpt.db[prev_hash] = prev_node
	}
	delete(mpt.db, cur_hash)
}

//==============================GET INSERT DELETE==========================================
// Get method to get value by key
func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	if key == "" {
		return "", errors.New("path_not_found")
	}
	if mpt.root == "" {
		return "", errors.New("path_not_found")
	}
	path := string2hexarray(key)
	ret := mpt.GetHelper(mpt.root, path)
	if ret != "" {
		return ret, nil
	}
	return "", errors.New("path_not_found")
}

// helper function for get
func (mpt *MerklePatriciaTrie) GetHelper(key string, path []uint8) string {
	// 0: Null, 1: Branch, 2: Ext or Leaf
	cur_node := mpt.db[key]
	switch cur_node.node_type {
	case 1: // Branch
		if len(path) == 0 {
			return cur_node.branch_value[16]
		}
		if cur_node.branch_value[path[0]] == "" {
			return ""
		}
		next_key := cur_node.branch_value[path[0]]
		return mpt.GetHelper(next_key, path[1:])

	case 2: // Ext or Leaf
		prefix := cur_node.flag_value.encoded_prefix
		decoded_prefix := compact_decode(prefix)
		if is_leaf(prefix) { //Leaf
			if eq_len(decoded_prefix, path) != len(path) {
				return ""
			} else {
				return cur_node.flag_value.value
			}
		} else { //Ext
			decode_prefix_len := len(decoded_prefix)
			if decode_prefix_len > len(path) || eq_len(decoded_prefix, path) != decode_prefix_len {
				fmt.Println("I failed at EXT", decoded_prefix, path)
				return ""
			} else {
				next_key := cur_node.flag_value.value
				return mpt.GetHelper(next_key, path[decode_prefix_len:])
			}
		}
	default:
		return ""
	}
}
func (mpt *MerklePatriciaTrie) GetJsonString() string {
	b, err := json.Marshal(mpt.plain)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// insert new key and value into mpt
func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	path := string2hexarray(key)
	if len(mpt.db) == 0 || mpt.root == "" {
		encoded := compact_encode(append(path, 16))
		flag := Flag_value{encoded_prefix: encoded, value: new_value}
		node := Node{node_type: 2, flag_value: flag}
		hashed := node.hash_node()
		mpt.db[hashed] = node
		mpt.root = hashed
	} else {
		mpt.InsertHelper(path, new_value, mpt.root, "")
	}
	mpt.plain[key] = new_value
}

// helper function for insert, pass the previous hash, current hash, path and value
func (mpt *MerklePatriciaTrie) InsertHelper(path []uint8, new_value string, cur_hash string, prev_hash string) string {
	node := mpt.db[cur_hash]
	switch node.node_type {
	case 1: // Branch
		if len(path) == 0 {
			node.branch_value[16] = new_value
		} else if node.branch_value[path[0]] == "" {
			node.branch_value[path[0]] = mpt.newLeafExtNode(path[1:], new_value, true)
		} else {
			node.branch_value[path[0]] = mpt.InsertHelper(path[1:], new_value, node.branch_value[path[0]], cur_hash)
		}
	case 2: // Ext or Leaf
		decoded := compact_decode(node.flag_value.encoded_prefix)
		eq_part, re_prefix, re_path := path_compare(path, decoded)
		leaf := is_leaf(node.flag_value.encoded_prefix)
		if len(re_prefix) == 0 && len(re_path) == 0 { // path == prefix are the same
			if leaf { //是leaf 就直接更新
				node.flag_value.value = new_value
			} else { //是ext, update the next level branch
				mpt.InsertHelper(re_path, new_value, node.flag_value.value, cur_hash)
				node = mpt.db[cur_hash]
			}
		} else if len(eq_part) == 0 {
			branch_value := [17]string{}
			// extension node with 1 element
			if len(re_path) != 0 {
				new_node_hash := mpt.newLeafExtNode(re_path[1:], new_value, true)
				branch_value[re_path[0]] = new_node_hash
			} else {
				branch_value[16] = new_value
			}

			if len(re_prefix) > 1 || leaf {
				if len(re_prefix) == 0 && leaf {
					branch_value[16] = node.flag_value.value
				} else {
					left_node_hash := mpt.newLeafExtNode(re_prefix[1:], node.flag_value.value, leaf)
					branch_value[re_prefix[0]] = left_node_hash
				}
			} else {
				branch_value[re_prefix[0]] = node.flag_value.value
			}

			node.node_type = 1
			node.branch_value = branch_value
		} else if len(re_prefix) == 0 && !leaf { // 把remaing的path 弄一个新的 branch插到ext下面的branch node 里面
			node.flag_value.value = mpt.InsertHelper(re_path, new_value, node.flag_value.value, cur_hash)
		} else { // re_prefix 依然存在 aab, aac => (aa, b, c)
			branch_value := [17]string{}
			if len(re_path) != 0 {
				new_node_hash := mpt.newLeafExtNode(re_path[1:], new_value, true)
				branch_value[re_path[0]] = new_node_hash
			} else {
				branch_value[16] = new_value
			}
			if len(re_prefix) != 0 {
				if !leaf && len(re_prefix) == 1 {
					branch_value[re_prefix[0]] = node.flag_value.value
				} else {
					left_node_hash := mpt.newLeafExtNode(re_prefix[1:], node.flag_value.value, leaf)
					branch_value[re_prefix[0]] = left_node_hash
				}
			} else { // is leaf
				branch_value[16] = node.flag_value.value
			}
			node.flag_value.encoded_prefix = compact_encode(eq_part) // change original leaf node to ext
			new_branch_hash := mpt.newBranchNode(branch_value)
			node.flag_value.value = new_branch_hash
		}
	}
	mpt.db[cur_hash] = node
	new_hash := node.hash_node()
	mpt.db[new_hash] = node
	mpt.UpdateHash(new_hash, cur_hash, prev_hash)
	return new_hash
}

// delete a node from tree
func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	path := string2hexarray(key)
	_, error := mpt.Get(key)
	if error != nil {
		return "", errors.New("path_not_found")
	}
	if len(mpt.db) == 1 {
		delete(mpt.db, mpt.root)
		mpt.root = ""
		return "", nil
	}
	mpt.DeleteHelper(path, mpt.root, "")
	delete(mpt.plain, key)
	return "", nil
}

// delete helper function
func (mpt *MerklePatriciaTrie) DeleteHelper(path []uint8, cur_hash string, prev_hash string) {
	node := mpt.db[cur_hash]
	deleted := false
	leaf := is_leaf(node.flag_value.encoded_prefix)
	switch node.node_type {
	case 1: //branch && deal with all leaf node inside it
		if len(path) == 0 {
			node.branch_value[16] = ""
			mpt.db[cur_hash] = node
		} else {
			mpt.DeleteHelper(path[1:], node.branch_value[path[0]], cur_hash)
			node = mpt.db[cur_hash]
		}
		branch_size := mpt.branchSize(cur_hash)
		if branch_size < 2 {
			prev_node := mpt.db[prev_hash]
			next_hash := ""

			if prev_node.node_type == 2 { //prev_node is ext, cur node deleted
				if node.branch_value[16] != "" {
					prev_node.flag_value.value = node.branch_value[16]
					decoded_prev_prefix := compact_decode(prev_node.flag_value.encoded_prefix)
					prev_node.flag_value.encoded_prefix = compact_encode(append(decoded_prev_prefix, 16))
				} else {
					i := 0
					for ; i < 16; i++ {
						if node.branch_value[i] != "" {
							next_hash = node.branch_value[i]
							break
						}
					}
					next_node := mpt.db[next_hash]
					decoded_prev_prefix := compact_decode(prev_node.flag_value.encoded_prefix)
					decoded_new_prev_prefix := append(decoded_prev_prefix, uint8(i))
					if next_node.node_type != 1 {
						decoded_next_prefix := compact_decode(next_node.flag_value.encoded_prefix)
						decoded_new_prev_prefix = append(decoded_new_prev_prefix, decoded_next_prefix...)
						if is_leaf(next_node.flag_value.encoded_prefix) {
							decoded_new_prev_prefix = append(decoded_new_prev_prefix, 16)
						}
						delete(mpt.db, next_hash)
						prev_node.flag_value.value = next_node.flag_value.value
					} else {
						prev_node.flag_value.value = next_hash
					}
					prev_node.flag_value.encoded_prefix = compact_encode(decoded_new_prev_prefix)
				}
				mpt.db[prev_hash] = prev_node
				delete(mpt.db, cur_hash)
				deleted = true
			} else { //prev_node is branch, cur node change to leaf node or ext
				node.node_type = 2
				flag_value := Flag_value{}
				node.flag_value = flag_value

				if node.branch_value[16] != "" { // branch to leaf node without path, 边上一个ext 已经删了
					node.flag_value.value = node.branch_value[16]
					node.flag_value.encoded_prefix = compact_encode([]uint8{16})
				} else { //
					i := 0
					for ; i < 16; i++ {
						if node.branch_value[i] != "" {
							next_hash = node.branch_value[i]
							break
						}
					}
					next_node := mpt.db[next_hash]
					new_prefix := []uint8{uint8(i)}
					if next_node.node_type == 1 { // br -> br -> br = br -> ex -> br
						node.flag_value.value = next_hash
					} else { // br -> br -> ex/leaf
						decoded_next_prefix := compact_decode(next_node.flag_value.encoded_prefix) // 你坏啦。。。。。。
						new_prefix = append(new_prefix, decoded_next_prefix...)
						if is_leaf(next_node.flag_value.encoded_prefix) {
							new_prefix = append(new_prefix, uint8(16))
						}
						node.flag_value.value = next_node.flag_value.value
						delete(mpt.db, next_hash)
					}
					node.flag_value.encoded_prefix = compact_encode(new_prefix)
				}
			}
		}
	case 2: // ext or leaf
		prefix := compact_decode(node.flag_value.encoded_prefix)
		if leaf && len(path) == eq_len(path, prefix) {
			prev_node := mpt.db[prev_hash]
			for i := 0; i < 16; i++ {
				if prev_node.branch_value[i] == cur_hash {
					prev_node.branch_value[i] = ""
					mpt.db[prev_hash] = prev_node
					deleted = true
					break
				}
			}
		} else {
			mpt.DeleteHelper(path[len(prefix):], node.flag_value.value, cur_hash)
			node = mpt.db[cur_hash]
		}
	}
	new_hash := node.hash_node()
	if !deleted {
		mpt.db[new_hash] = node
		mpt.UpdateHash(new_hash, cur_hash, prev_hash)
	}
}

func compact_decode(encoded_arr []uint8) []uint8 {
	// TODO
	ret := []uint8{}
	size := len(encoded_arr)
	for i := 0; i < size; i++ {
		p := encoded_arr[i] / 16
		q := encoded_arr[i] % 16
		ret = append(ret, p)
		ret = append(ret, q)
	}
	if ret[1] == 0 && (ret[0] == 0 || ret[0] == 2) {
		return ret[2:]
	}
	return ret[1:]
}

func compact_encode(hex_array []uint8) []uint8 {
	if len(hex_array) == 0 {
		return []uint8{0, 0}
	}
	size := len(hex_array)
	ret := []uint8{}
	term := 0
	if hex_array[size-1] == 16 {
		term = 1
	}
	oddlen := (size - term) % 2
	flags := 2*term + oddlen
	tmp := []uint8{uint8(flags)}
	if oddlen == 0 {
		tmp = append(tmp, 0)
	}

	encoded_array := append(tmp, hex_array[:size-term]...)
	for i := 0; i < len(encoded_array)-1; i += 2 {
		t := encoded_array[i]
		f := encoded_array[i+1]
		ret = append(ret, uint8(t*16+f))
	}

	if len(encoded_array)%2 == 1 {
		ret = append(ret, encoded_array[len(encoded_array)-1])
	}
	return ret
}
