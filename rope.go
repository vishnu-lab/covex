package main

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// type Node interface {
//Y 	String(builder strings.Builder)
//Y 	At(index int) (rune, error)
//Y 	Substring(start, end int) (string, error)

//Y 	Split(index int) (Node, Node, error)
//Y 	Concat(other Node) Node
//Y 	Slice(start, end int) (*Rope, error)
//Y  leaf(text string) *Node

//Y 	Insert(index int, text string) error
//Y 	Delete(start, end int) error
//N 	Replace(start, end int, text string) error

//N 	Find(search string, start int) int
//Y  	FindAll(search string) []int
// 	LineStart(lineNum int) int
// 	LineEnd(lineNum int) int

// 	IndexToLineCol(index int) (line, col int)
// 	LineColToIndex(line, col int) int
// }

type Node struct {
	isLeaf             bool
	charCount          int //this denotes the number of characters present in the node
	lineCount          int //this denotes the number of completed lines that is , the number of \n's present
	remainingCharCount int //this denotes the number of charcaters present after the last occurance of \n
	content            []byte
	left, right        *Node
}

func (node *Node) String(builder *strings.Builder) {

	if node.isLeaf {
		builder.Write(node.content)
		return
	}
	if node.left != nil {
		node.left.String(builder)
	}
	if node.right != nil {
		node.right.String(builder)
	}
	return
}

// this is 0-indexed
func (node *Node) At(lineNumber, colNumber int) (rune, error) {
	if lineNumber < 0 || colNumber < 0 || lineNumber > node.lineCount || (lineNumber == node.lineCount && colNumber >= node.remainingCharCount) {
		return -1, errors.New("Index is out of bounds")
	}
	if node.isLeaf {
		runes := []rune(string(node.content)) // convert to runes
		i := 0
		for lineNumber > 0 {
			if i >= len(runes) {
				return -1, errors.New("Index out of bounds, but i am not sure it came till here, there is some error in the source code")
			}
			if runes[i] == '\n' {
				lineNumber = lineNumber - 1
			}
			i = i + 1
		}
		ch := runes[i+colNumber]
		return ch, nil
	}
	if lineNumber < node.left.lineCount || (lineNumber == node.left.lineCount && colNumber < node.left.remainingCharCount) {
		return node.left.At(lineNumber, colNumber)
	}
	colNumberEq := colNumber
	if lineNumber == node.left.lineCount {
		colNumberEq = colNumberEq - node.left.remainingCharCount
	}
	return node.right.At(lineNumber-node.left.lineCount, colNumberEq)
}

// inclusive of start, exclusive of end
func (node *Node) Substring(startLineNum, startColNum, endLineNum, endColNum int) (string, error) {
	if startLineNum > node.lineCount || (startLineNum == node.lineCount && startColNum >= node.remainingCharCount) {
		return "", errors.New("Out of Bounds error")
	}

	if endLineNum > node.lineCount || (endLineNum == node.lineCount && endColNum > node.remainingCharCount) {
		return "", errors.New("Out of Bounds error")
	}

	if startLineNum > endLineNum || (startLineNum == endLineNum && startColNum >= endColNum) || startLineNum < 0 || startColNum < 0 || endLineNum < 0 || endColNum < 0 {
		return "", errors.New("Invalid input")
	}

	if node.isLeaf {
		runes := []rune(string(node.content))
		curLine := 0
		curCol := 0
		i := 0
		startIndex := -1
		endIndex := -1
		for i < len(runes) {
			if curLine == startLineNum && curCol == startColNum {
				startIndex = i
				break
			}
			if runes[i] == '\n' {
				curLine++
				curCol = 0
			} else {
				curCol++
			}
			i++
		}
		for i <= len(runes) {
			if curLine == endLineNum && curCol == endColNum {
				endIndex = i
				break
			}
			if i == len(runes) {
				break
			}
			if runes[i] == '\n' {
				curLine++
				curCol = 0
			} else {
				curCol++
			}
			i++
		}
		return string(runes[startIndex:endIndex]), nil
	}

	if endLineNum < node.left.lineCount || (endLineNum == node.left.lineCount && endColNum <= node.left.remainingCharCount) {
		leftString, _ := node.left.Substring(startLineNum, startColNum, endLineNum, endColNum)
		return leftString, nil
	}
	if startLineNum > node.left.lineCount || (startLineNum == node.left.lineCount && startColNum >= node.left.remainingCharCount) {
		startColNumEq := startColNum
		if startLineNum == node.left.lineCount {
			startColNumEq = startColNum - node.left.remainingCharCount
		}
		endColNumEq := endColNum
		if endLineNum == node.left.lineCount {
			endColNumEq = endColNum - node.left.remainingCharCount
		}
		rightString, _ := node.right.Substring(startLineNum-node.left.lineCount, startColNumEq, endLineNum-node.left.lineCount, endColNumEq)
		return rightString, nil
	}
	leftString, _ := node.left.Substring(startLineNum, startColNum, node.left.lineCount, node.left.remainingCharCount)
	var rightString string
	if endLineNum > node.left.lineCount {
		rightString, _ = node.right.Substring(0, 0, endLineNum-node.left.lineCount, endColNum)
	} else {
		rightString, _ = node.right.Substring(0, 0, endLineNum-node.left.lineCount, endColNum-node.left.remainingCharCount)
	}
	return leftString + rightString, nil
}

// it splits the rope to 2 ropes where the second starts from lineNum, colNum charcter ( 0 indexed line number and column number)
func (node *Node) Split(lineNum, colNum int) (*Node, *Node, error) {
	if lineNum > node.lineCount || (lineNum == node.lineCount && colNum >= node.remainingCharCount) || lineNum < 0 || colNum < 0 {
		return nil, nil, errors.New("Out of Bounds error")
	}

	if lineNum == 0 && colNum == 0 {
		emptyNode := &Node{
			isLeaf:             true,
			content:            []byte{},
			charCount:          0,
			lineCount:          0,
			remainingCharCount: 0,
		}
		return emptyNode, node, nil
	}
	if lineNum == node.lineCount && colNum == node.remainingCharCount {
		emptyNode := &Node{
			isLeaf:             true,
			content:            []byte{},
			charCount:          0,
			lineCount:          0,
			remainingCharCount: 0,
		}
		return node, emptyNode, nil
	}

	if node.isLeaf {
		runes := []rune(string(node.content))
		splitIndex := 0
		currentLine := 0
		currentCol := 0

		// Iterate rune by rune to find the split point
		for splitIndex < len(runes) {
			if currentLine == lineNum && currentCol == colNum {
				break
			}
			if runes[splitIndex] == '\n' {
				currentLine++
				currentCol = 0
			} else {
				currentCol++
			}
			splitIndex++
		}

		if splitIndex >= len(runes) {
			return nil, nil, errors.New("Index out of bounds")
		}
		leftContentString := string(runes[:splitIndex])
		rightContentString := string(runes[splitIndex:])

		leftNode := leaf(leftContentString)   // Reuse the 'leaf' constructor
		rightNode := leaf(rightContentString) // It correctly calculates all metadata

		return leftNode, rightNode, nil
	}

	if lineNum == node.left.lineCount && colNum == node.left.remainingCharCount {
		return node.left, node.right, nil
	}

	if lineNum < node.left.lineCount || (lineNum == node.left.lineCount && colNum < node.left.remainingCharCount) {
		leftSplit, rightSplit, err := node.left.Split(lineNum, colNum)
		if err != nil {
			return nil, nil, err
		}
		return leftSplit, rightSplit.Concat(node.right), nil
	} else {
		colNumEq := colNum
		if lineNum == node.left.lineCount {
			colNumEq = colNum - node.left.remainingCharCount
		}
		leftSplit, rightSplit, err := node.right.Split(lineNum-node.left.lineCount, colNumEq)
		if err != nil {
			return nil, nil, err
		}
		return node.left.Concat(leftSplit), rightSplit, nil
	}
}

func (node *Node) Concat(other *Node) *Node {
	if node == nil {
		return other
	}
	if other == nil {
		return node
	}

	// Handle empty nodes: if one is empty, return the other
	// An empty node has charCount == 0 and lineCount == 0
	if node.charCount == 0 {
		return other
	}
	if other.charCount == 0 {
		return node
	}

	// Calculate combined character count
	charCount := node.charCount + other.charCount
	lineCount := node.lineCount + other.lineCount

	// The remaining character count always comes from the rightmost node
	remainingCharCount := other.remainingCharCount

	if other.lineCount == 0 {
		remainingCharCount += node.remainingCharCount
	}

	// Create a new internal node
	return &Node{
		isLeaf:             false,
		charCount:          charCount,
		lineCount:          lineCount,
		remainingCharCount: remainingCharCount,
		left:               node,
		right:              other,
	}
}

// inclusive of start and exclusive of end
func (node *Node) Slice(startLineNum, startColNum, endLineNum, endColNum int) (*Node, error) {
	if startLineNum > node.lineCount || (startLineNum == node.lineCount && startColNum >= node.remainingCharCount) {
		return nil, errors.New("Out of Bounds error")
	}

	if endLineNum > node.lineCount || (endLineNum == node.lineCount && endColNum > node.remainingCharCount) {
		return nil, errors.New("Out of Bounds error")
	}

	if startLineNum > endLineNum || (startLineNum == endLineNum && startColNum >= endColNum) || startLineNum < 0 || startColNum < 0 || endLineNum < 0 || endColNum < 0 {
		return nil, errors.New("Invalid input")
	}

	if node.isLeaf {
		runes := []rune(string(node.content))
		curLine := 0
		curCol := 0
		i := 0
		startIndex := -1
		endIndex := -1
		for i < len(runes) {
			if curLine == startLineNum && curCol == startColNum {
				startIndex = i
				break
			}
			if runes[i] == '\n' {
				curLine++
				curCol = 0
			} else {
				curCol++
			}
			i++
		}
		for i <= len(runes) {
			if curLine == endLineNum && curCol == endColNum {
				endIndex = i
				break
			}
			if i == len(runes) {
				break
			}
			if runes[i] == '\n' {
				curLine++
				curCol = 0
			} else {
				curCol++
			}
			i++
		}
		return leaf(string(runes[startIndex:endIndex])), nil
	}

	if endLineNum < node.left.lineCount || (endLineNum == node.left.lineCount && endColNum <= node.left.remainingCharCount) {
		leftSlice, _ := node.left.Slice(startLineNum, startColNum, endLineNum, endColNum)
		return leftSlice, nil
	}
	if startLineNum > node.left.lineCount || (startLineNum == node.left.lineCount && startColNum >= node.left.remainingCharCount) {
		startColNumEq := startColNum
		if startLineNum == node.left.lineCount {
			startColNumEq = startColNum - node.left.remainingCharCount
		}
		endColNumEq := endColNum
		if endLineNum == node.left.lineCount {
			endColNumEq = endColNum - node.left.remainingCharCount
		}
		rightSlice, _ := node.right.Slice(startLineNum-node.left.lineCount, startColNumEq, endLineNum-node.left.lineCount, endColNumEq)
		return rightSlice, nil
	}
	leftSlice, _ := node.left.Slice(startLineNum, startColNum, node.left.lineCount, node.left.remainingCharCount)
	var rightSlice *Node
	if endLineNum > node.left.lineCount {
		rightSlice, _ = node.right.Slice(0, 0, endLineNum-node.left.lineCount, endColNum)
	} else {
		rightSlice, _ = node.right.Slice(0, 0, endLineNum-node.left.lineCount, endColNum-node.left.remainingCharCount)
	}
	return leftSlice.Concat(rightSlice), nil
}

func leaf(text string) *Node {
	remainingCharCount := 0
	lineCount := 0
	for _, r := range text {
		if r == '\n' {
			lineCount++
			remainingCharCount = 0
		} else {
			remainingCharCount++
		}
	}
	return &Node{
		isLeaf:             true,
		content:            []byte(text),
		charCount:          utf8.RuneCountInString(text),
		lineCount:          lineCount,
		remainingCharCount: remainingCharCount,
	}
}

// inserts at given position, the first char of text will be placed at lineNum, colNum
func (node *Node) Insert(lineNum, colNum int, text string) (*Node, error) {
	if lineNum > node.lineCount || (lineNum == node.lineCount && colNum > node.remainingCharCount) || lineNum < 0 || colNum < 0 {
		return nil, errors.New("Out of Bounds error")
	}
	if lineNum == 0 && colNum == 0 {
		return leaf(text).Concat(node), nil
	}
	if lineNum == node.lineCount && colNum == node.remainingCharCount {
		return node.Concat(leaf(text)), nil
	}

	leftSplit, rightSplit, err := node.Split(lineNum, colNum)
	if err != nil {
		return nil, err
	}
	return leftSplit.Concat(leaf(text)).Concat(rightSplit), nil
}

// inclusive of start and exclusive of end
func (node *Node) Delete(startLineNum, startColNum, endLineNum, endColNum int) (*Node, error) {
	if startLineNum > node.lineCount || (startLineNum == node.lineCount && startColNum >= node.remainingCharCount) {
		return nil, errors.New("Out of Bounds error")
	}
	if endLineNum > node.lineCount || (endLineNum == node.lineCount && endColNum > node.remainingCharCount) {
		return nil, errors.New("Out of Bounds error")
	}
	if startLineNum > endLineNum || (startLineNum == endLineNum && startColNum >= endColNum) || startLineNum < 0 || startColNum < 0 || endLineNum < 0 || endColNum < 0 {
		return nil, errors.New("Invalid input")
	}
	if startLineNum == endLineNum && startColNum == endColNum {
		return node, nil
	}

	leftSplit, rightSplit, err := node.Split(startLineNum, startColNum)
	if err != nil {
		return nil, err
	}
	endColNumEq := endColNum
	if endLineNum == leftSplit.lineCount {
		endColNumEq = endColNumEq - leftSplit.remainingCharCount
	}
	_, rightSplitreq, _ := rightSplit.Split(endLineNum-leftSplit.lineCount, endColNumEq)
	return leftSplit.Concat(rightSplitreq), nil
}

func (node *Node) getIterator() (iterator *NodeIter) {
	iterator = &NodeIter{
		stack:         []*Node{},
		buffer:        []rune{},
		indexOfBuffer: 0,
	}
	iterator.push(node)
	return
}

// finds all the positions that matches the string
func (node *Node) FindAll(search string) []Position {
	positions := []Position{}
	if len(search) == 0 {
		return positions
	}
	searchPattern := []rune(search)
	curPattern := []rune{}
	iterator := node.getIterator()
	curPositionMatches := false
	for i := 0; i < len(searchPattern); i++ {
		nextRune, err := iterator.NextRune()
		if err != nil {
			return positions
		}
		curPattern = append(curPattern, nextRune)
	}
	currentPosition := Position{Line: 0, Column: 0}
	for true {
		curPositionMatches = true
		for i := 0; i < len(searchPattern); i++ {
			if curPattern[i] != searchPattern[i] {
				curPositionMatches = false
				break
			}
		}
		if curPositionMatches {
			positions = append(positions, currentPosition)
		}
		if curPattern[0] == '\n' {
			currentPosition.Line++
			currentPosition.Column = 0
		} else {
			currentPosition.Column++
		}
		curPattern = curPattern[1:]
		nextRune, err := iterator.NextRune()
		if err != nil {
			return positions
		}
		curPattern = append(curPattern, nextRune)
	}
	return positions
}

type Position struct {
	Line   int
	Column int
}

type NodeIter struct {
	stack         []*Node
	buffer        []rune
	indexOfBuffer int
}

func (iter *NodeIter) NextRune() (rune, error) {
	for iter.indexOfBuffer >= len(iter.buffer) {
		err := iter.loadNextLeaf()
		if err != nil {
			return 'a', err
		}
	}
	if iter.indexOfBuffer < len(iter.buffer) {
		nextRune := iter.buffer[iter.indexOfBuffer]
		iter.indexOfBuffer++
		return nextRune, nil
	} else {
		return 'a', errors.New("Index out of bounds")
	}
}

func (iter *NodeIter) loadNextLeaf() error {
	if len(iter.stack) == 0 {
		return errors.New("Index out of bounds")
	}
	for len(iter.stack) != 0 {
		index := len(iter.stack) - 1
		element := iter.stack[index]
		iter.stack = iter.stack[:index]
		if element.isLeaf {
			iter.buffer = []rune(string(element.content))
			iter.indexOfBuffer = 0
			return nil
		}
		iter.push(element.right)
	}
	return errors.New("Index out of Bounds")
}

func (iter *NodeIter) push(node *Node) {
	if node == nil {
		return
	}
	iter.stack = append(iter.stack, node)
	iter.push(node.left)
}
