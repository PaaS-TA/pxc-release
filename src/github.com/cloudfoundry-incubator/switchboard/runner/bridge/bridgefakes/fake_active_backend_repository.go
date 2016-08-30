// This file was generated by counterfeiter
package bridgefakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/switchboard/domain"
	"github.com/cloudfoundry-incubator/switchboard/runner/bridge"
)

type FakeActiveBackendRepository struct {
	ActiveStub        func() domain.Backend
	activeMutex       sync.RWMutex
	activeArgsForCall []struct{}
	activeReturns     struct {
		result1 domain.Backend
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeActiveBackendRepository) Active() domain.Backend {
	fake.activeMutex.Lock()
	fake.activeArgsForCall = append(fake.activeArgsForCall, struct{}{})
	fake.recordInvocation("Active", []interface{}{})
	fake.activeMutex.Unlock()
	if fake.ActiveStub != nil {
		return fake.ActiveStub()
	} else {
		return fake.activeReturns.result1
	}
}

func (fake *FakeActiveBackendRepository) ActiveCallCount() int {
	fake.activeMutex.RLock()
	defer fake.activeMutex.RUnlock()
	return len(fake.activeArgsForCall)
}

func (fake *FakeActiveBackendRepository) ActiveReturns(result1 domain.Backend) {
	fake.ActiveStub = nil
	fake.activeReturns = struct {
		result1 domain.Backend
	}{result1}
}

func (fake *FakeActiveBackendRepository) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.activeMutex.RLock()
	defer fake.activeMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeActiveBackendRepository) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ bridge.ActiveBackendRepository = new(FakeActiveBackendRepository)
