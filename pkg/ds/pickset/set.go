package pickset

import (
	"github.com/pkg/errors"
	"github.com/zyedidia/generic/list"
	"sync"
	"time"
)

// PickableSet is a data structure that provides the basic functionalities of a set and meanwhile provides a round-robin
// picking mechanism and the ability to block an element being picked permanently or in a given duration.
type PickableSet[T comparable] struct {
	list       *list.List[T]
	searchMap  map[T]*list.Node[T]
	blockedSet map[T]struct{}
	mut        *sync.RWMutex
}

func NewPickableSet[T comparable]() *PickableSet[T] {
	return &PickableSet[T]{
		list:       list.New[T](),
		searchMap:  make(map[T]*list.Node[T]),
		blockedSet: make(map[T]struct{}),
		mut:        &sync.RWMutex{},
	}
}

func (s *PickableSet[T]) Len() int {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.len()
}

func (s *PickableSet[T]) len() int {
	return len(s.searchMap)
}

// Add returns true if the element is added, and false if the element already exists and therefore not added.
func (s *PickableSet[T]) Add(key T) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.add(key)
}

func (s *PickableSet[T]) add(key T) bool {
	if _, ok := s.searchMap[key]; ok {
		return false
	}
	newNode := &list.Node[T]{Value: key}
	s.list.PushBackNode(newNode)
	s.searchMap[key] = newNode
	return true
}

// Remove returns true if the element is removed, and false if the element does not exist and therefore not removed.
func (s *PickableSet[T]) Remove(key T) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.remove(key)
}

func (s *PickableSet[T]) remove(key T) bool {
	node, ok := s.searchMap[key]
	if !ok {
		return false
	}
	delete(s.searchMap, key)
	s.list.Remove(node)
	if _, ok := s.blockedSet[key]; ok {
		delete(s.blockedSet, key)
	}
	return true
}

func (s *PickableSet[T]) Has(key T) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.has(key)
}

func (s *PickableSet[T]) has(key T) bool {
	_, ok := s.searchMap[key]
	return ok
}

func (s *PickableSet[T]) HasUnblocked(key T) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.hasUnblocked(key)
}

func (s *PickableSet[T]) hasUnblocked(key T) bool {
	_, searchOk := s.searchMap[key]
	_, blockOk := s.blockedSet[key]
	return searchOk && !blockOk
}

// Block returns true if the element is blocked, or false if element is already blocked or element does not exist.
func (s *PickableSet[T]) Block(key T) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.block(key)
}

func (s *PickableSet[T]) block(key T) bool {
	if _, searchOk := s.searchMap[key]; !searchOk {
		return false
	}
	if _, blockOk := s.blockedSet[key]; blockOk {
		return false
	}
	s.blockedSet[key] = struct{}{}
	return true
}

// Unblock returns true if the element is unblocked, or false if element is not blocked or element does not exist.
func (s *PickableSet[T]) Unblock(key T) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.unblock(key)
}

func (s *PickableSet[T]) unblock(key T) bool {
	if _, searchOk := s.searchMap[key]; !searchOk {
		return false
	}
	if _, blockOk := s.blockedSet[key]; !blockOk {
		return false
	}
	delete(s.blockedSet, key)
	return true
}

func (s *PickableSet[T]) IsBlocked(key T) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.isBlocked(key)
}

func (s *PickableSet[T]) isBlocked(key T) bool {
	_, ok := s.blockedSet[key]
	return ok
}

func (s *PickableSet[T]) BlockForDuration(key T, duration time.Duration) bool {
	if s.IsBlocked(key) {
		return false
	}
	s.Block(key)
	go func() {
		afterCh := time.After(duration)
		<-afterCh
		s.Unblock(key)
	}()
	return true
}

type Picker[T comparable] struct {
	ps       *PickableSet[T]
	prevPick *list.Node[T]
	mut      sync.RWMutex
	all      bool
}

func NewPicker[T comparable](ps *PickableSet[T]) *Picker[T] {
	return &Picker[T]{
		ps:       ps,
		prevPick: ps.list.Front,
		mut:      sync.RWMutex{},
		all:      false,
	}
}

func NewAllPicker[T comparable](ps *PickableSet[T]) *Picker[T] {
	return &Picker[T]{
		ps:       ps,
		prevPick: ps.list.Back,
		mut:      sync.RWMutex{},
		all:      true,
	}
}

func (p *Picker[T]) hasValue(key T) bool {
	if p.all {
		ret := p.ps.has(key)
		// should not ever happen
		if ret == false {
			panic(ErrListIsInvalid)
		}
		return ret
	} else {
		return p.ps.hasUnblocked(key)
	}
}

// Pick picks the next unblocked value of the PickableSet
func (p *Picker[T]) Pick() (ret T, err error) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.ps.mut.RLock()
	defer p.ps.mut.RUnlock()
	if p.ps.list.Front == nil && p.ps.list.Back == nil {
		return ret, ErrListIsEmpty
	}
	if p.ps.list.Front == nil || p.ps.list.Back == nil {
		return ret, ErrListIsInvalid
	}
	if p.ps.len() == 0 {
		return ret, ErrSetIsEmpty
	}
	if len(p.ps.searchMap) == len(p.ps.blockedSet) {
		return ret, ErrNoElementAvailableForPicking
	}
	if isDanglingNode(p.prevPick) {
		p.prevPick = p.ps.list.Back
	}

	currentPick := p.prevPick.Next

	if currentPick == nil {
		currentPick = p.ps.list.Front
	}

	arrivedHead := false
	arrivedTail := true

	maxIter := p.ps.Len() * 3

	iterCount := 0

	for iterCount < maxIter {
		iterCount++
		if p.hasValue(currentPick.Value) {
			p.prevPick = currentPick
			return currentPick.Value, nil
		}
		if currentPick.Prev == nil {
			arrivedHead = true
		}
		if currentPick.Next == nil {
			arrivedTail = true
		}
		if arrivedHead && arrivedTail {
			return *new(T), ErrNoElementAvailableForPicking
		}
		if currentPick.Next == nil {
			currentPick = p.ps.list.Front
		} else {
			currentPick = currentPick.Next
		}
	}

	return ret, ErrMaxIteration

}

func isDanglingNode[T comparable](node *list.Node[T]) bool {
	return node == nil || node.Prev == nil && node.Next == nil
}

var ErrNoElementAvailableForPicking = errors.New("no element available for picking")

var ErrSetIsEmpty = errors.New("set is empty")

var ErrListIsEmpty = errors.New("list is empty")

var ErrListIsInvalid = errors.New("list integrity failure")

var ErrMaxIteration = errors.New("PickableSet is invalid, cannot find an available after going through maximum iterations")
