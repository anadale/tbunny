package model

import "sync"

const (
	stackPush stackAction = 1 << iota
	stackPop
)

type stackAction int

type ViewStackListener interface {
	StackPushed(View)
	StackPopped(old, new View)
	StackTop(View)
}

type ViewStack struct {
	components []View
	listeners  []ViewStackListener
	mx         sync.RWMutex
}

func NewViewStack() *ViewStack {
	return &ViewStack{}
}

func (v *ViewStack) AddListener(l ViewStackListener) {
	v.listeners = append(v.listeners, l)

	if !v.Empty() {
		l.StackTop(v.Top())
	}
}

func (v *ViewStack) RemoveListener(l ViewStackListener) {
	for i, l2 := range v.listeners {
		if l2 == l {
			v.listeners = append(v.listeners[:i], v.listeners[i+1:]...)
			return
		}
	}
}

func (v *ViewStack) Push(c View) {
	v.mx.Lock()
	v.components = append(v.components, c)
	v.mx.Unlock()
	v.notify(stackPush, c)
}

func (v *ViewStack) Pop() (View, bool) {
	if v.Empty() {
		return nil, false
	}

	var c View

	v.mx.Lock()
	c = v.components[len(v.components)-1]
	v.components = v.components[:len(v.components)-1]
	v.mx.Unlock()

	v.notify(stackPop, c)

	return c, true
}

func (v *ViewStack) Clear() {
	for range v.components {
		v.Pop()
	}
}

func (v *ViewStack) Empty() bool {
	v.mx.RLock()
	defer v.mx.RUnlock()

	return len(v.components) == 0
}

func (v *ViewStack) Last() bool {
	v.mx.RLock()
	defer v.mx.RUnlock()

	return len(v.components) == 1
}

func (v *ViewStack) Top() View {
	if v.Empty() {
		return nil
	}

	v.mx.RLock()
	defer v.mx.RUnlock()

	return v.components[len(v.components)-1]
}

func (v *ViewStack) CollectNames() []string {
	v.mx.RLock()
	defer v.mx.RUnlock()

	names := make([]string, len(v.components))
	for i, c := range v.components {
		names[i] = c.Name()
	}

	return names
}

func (v *ViewStack) notify(action stackAction, c View) {
	for _, l := range v.listeners {
		switch action {
		case stackPush:
			l.StackPushed(c)
		case stackPop:
			l.StackPopped(c, v.Top())
		}
	}
}
