// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/seattle-beach/cf-cli-rds-plugin/api"
)

type FakeRDSService struct {
	DescribeDBSubnetGroupsStub        func(input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error)
	describeDBSubnetGroupsMutex       sync.RWMutex
	describeDBSubnetGroupsArgsForCall []struct {
		input *rds.DescribeDBSubnetGroupsInput
	}
	describeDBSubnetGroupsReturns struct {
		result1 *rds.DescribeDBSubnetGroupsOutput
		result2 error
	}
	describeDBSubnetGroupsReturnsOnCall map[int]struct {
		result1 *rds.DescribeDBSubnetGroupsOutput
		result2 error
	}
	CreateDBInstanceStub        func(input *rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, error)
	createDBInstanceMutex       sync.RWMutex
	createDBInstanceArgsForCall []struct {
		input *rds.CreateDBInstanceInput
	}
	createDBInstanceReturns struct {
		result1 *rds.CreateDBInstanceOutput
		result2 error
	}
	createDBInstanceReturnsOnCall map[int]struct {
		result1 *rds.CreateDBInstanceOutput
		result2 error
	}
	DescribeDBInstancesStub        func(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
	describeDBInstancesMutex       sync.RWMutex
	describeDBInstancesArgsForCall []struct {
		input *rds.DescribeDBInstancesInput
	}
	describeDBInstancesReturns struct {
		result1 *rds.DescribeDBInstancesOutput
		result2 error
	}
	describeDBInstancesReturnsOnCall map[int]struct {
		result1 *rds.DescribeDBInstancesOutput
		result2 error
	}
	ModifyDBInstanceStub        func(input *rds.ModifyDBInstanceInput) (*rds.ModifyDBInstanceOutput, error)
	modifyDBInstanceMutex       sync.RWMutex
	modifyDBInstanceArgsForCall []struct {
		input *rds.ModifyDBInstanceInput
	}
	modifyDBInstanceReturns struct {
		result1 *rds.ModifyDBInstanceOutput
		result2 error
	}
	modifyDBInstanceReturnsOnCall map[int]struct {
		result1 *rds.ModifyDBInstanceOutput
		result2 error
	}
	WaitUntilDBInstanceAvailableStub        func(input *rds.DescribeDBInstancesInput) error
	waitUntilDBInstanceAvailableMutex       sync.RWMutex
	waitUntilDBInstanceAvailableArgsForCall []struct {
		input *rds.DescribeDBInstancesInput
	}
	waitUntilDBInstanceAvailableReturns struct {
		result1 error
	}
	waitUntilDBInstanceAvailableReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRDSService) DescribeDBSubnetGroups(input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error) {
	fake.describeDBSubnetGroupsMutex.Lock()
	ret, specificReturn := fake.describeDBSubnetGroupsReturnsOnCall[len(fake.describeDBSubnetGroupsArgsForCall)]
	fake.describeDBSubnetGroupsArgsForCall = append(fake.describeDBSubnetGroupsArgsForCall, struct {
		input *rds.DescribeDBSubnetGroupsInput
	}{input})
	fake.recordInvocation("DescribeDBSubnetGroups", []interface{}{input})
	fake.describeDBSubnetGroupsMutex.Unlock()
	if fake.DescribeDBSubnetGroupsStub != nil {
		return fake.DescribeDBSubnetGroupsStub(input)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.describeDBSubnetGroupsReturns.result1, fake.describeDBSubnetGroupsReturns.result2
}

func (fake *FakeRDSService) DescribeDBSubnetGroupsCallCount() int {
	fake.describeDBSubnetGroupsMutex.RLock()
	defer fake.describeDBSubnetGroupsMutex.RUnlock()
	return len(fake.describeDBSubnetGroupsArgsForCall)
}

func (fake *FakeRDSService) DescribeDBSubnetGroupsArgsForCall(i int) *rds.DescribeDBSubnetGroupsInput {
	fake.describeDBSubnetGroupsMutex.RLock()
	defer fake.describeDBSubnetGroupsMutex.RUnlock()
	return fake.describeDBSubnetGroupsArgsForCall[i].input
}

func (fake *FakeRDSService) DescribeDBSubnetGroupsReturns(result1 *rds.DescribeDBSubnetGroupsOutput, result2 error) {
	fake.DescribeDBSubnetGroupsStub = nil
	fake.describeDBSubnetGroupsReturns = struct {
		result1 *rds.DescribeDBSubnetGroupsOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) DescribeDBSubnetGroupsReturnsOnCall(i int, result1 *rds.DescribeDBSubnetGroupsOutput, result2 error) {
	fake.DescribeDBSubnetGroupsStub = nil
	if fake.describeDBSubnetGroupsReturnsOnCall == nil {
		fake.describeDBSubnetGroupsReturnsOnCall = make(map[int]struct {
			result1 *rds.DescribeDBSubnetGroupsOutput
			result2 error
		})
	}
	fake.describeDBSubnetGroupsReturnsOnCall[i] = struct {
		result1 *rds.DescribeDBSubnetGroupsOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) CreateDBInstance(input *rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, error) {
	fake.createDBInstanceMutex.Lock()
	ret, specificReturn := fake.createDBInstanceReturnsOnCall[len(fake.createDBInstanceArgsForCall)]
	fake.createDBInstanceArgsForCall = append(fake.createDBInstanceArgsForCall, struct {
		input *rds.CreateDBInstanceInput
	}{input})
	fake.recordInvocation("CreateDBInstance", []interface{}{input})
	fake.createDBInstanceMutex.Unlock()
	if fake.CreateDBInstanceStub != nil {
		return fake.CreateDBInstanceStub(input)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.createDBInstanceReturns.result1, fake.createDBInstanceReturns.result2
}

func (fake *FakeRDSService) CreateDBInstanceCallCount() int {
	fake.createDBInstanceMutex.RLock()
	defer fake.createDBInstanceMutex.RUnlock()
	return len(fake.createDBInstanceArgsForCall)
}

func (fake *FakeRDSService) CreateDBInstanceArgsForCall(i int) *rds.CreateDBInstanceInput {
	fake.createDBInstanceMutex.RLock()
	defer fake.createDBInstanceMutex.RUnlock()
	return fake.createDBInstanceArgsForCall[i].input
}

func (fake *FakeRDSService) CreateDBInstanceReturns(result1 *rds.CreateDBInstanceOutput, result2 error) {
	fake.CreateDBInstanceStub = nil
	fake.createDBInstanceReturns = struct {
		result1 *rds.CreateDBInstanceOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) CreateDBInstanceReturnsOnCall(i int, result1 *rds.CreateDBInstanceOutput, result2 error) {
	fake.CreateDBInstanceStub = nil
	if fake.createDBInstanceReturnsOnCall == nil {
		fake.createDBInstanceReturnsOnCall = make(map[int]struct {
			result1 *rds.CreateDBInstanceOutput
			result2 error
		})
	}
	fake.createDBInstanceReturnsOnCall[i] = struct {
		result1 *rds.CreateDBInstanceOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
	fake.describeDBInstancesMutex.Lock()
	ret, specificReturn := fake.describeDBInstancesReturnsOnCall[len(fake.describeDBInstancesArgsForCall)]
	fake.describeDBInstancesArgsForCall = append(fake.describeDBInstancesArgsForCall, struct {
		input *rds.DescribeDBInstancesInput
	}{input})
	fake.recordInvocation("DescribeDBInstances", []interface{}{input})
	fake.describeDBInstancesMutex.Unlock()
	if fake.DescribeDBInstancesStub != nil {
		return fake.DescribeDBInstancesStub(input)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.describeDBInstancesReturns.result1, fake.describeDBInstancesReturns.result2
}

func (fake *FakeRDSService) DescribeDBInstancesCallCount() int {
	fake.describeDBInstancesMutex.RLock()
	defer fake.describeDBInstancesMutex.RUnlock()
	return len(fake.describeDBInstancesArgsForCall)
}

func (fake *FakeRDSService) DescribeDBInstancesArgsForCall(i int) *rds.DescribeDBInstancesInput {
	fake.describeDBInstancesMutex.RLock()
	defer fake.describeDBInstancesMutex.RUnlock()
	return fake.describeDBInstancesArgsForCall[i].input
}

func (fake *FakeRDSService) DescribeDBInstancesReturns(result1 *rds.DescribeDBInstancesOutput, result2 error) {
	fake.DescribeDBInstancesStub = nil
	fake.describeDBInstancesReturns = struct {
		result1 *rds.DescribeDBInstancesOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) DescribeDBInstancesReturnsOnCall(i int, result1 *rds.DescribeDBInstancesOutput, result2 error) {
	fake.DescribeDBInstancesStub = nil
	if fake.describeDBInstancesReturnsOnCall == nil {
		fake.describeDBInstancesReturnsOnCall = make(map[int]struct {
			result1 *rds.DescribeDBInstancesOutput
			result2 error
		})
	}
	fake.describeDBInstancesReturnsOnCall[i] = struct {
		result1 *rds.DescribeDBInstancesOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) ModifyDBInstance(input *rds.ModifyDBInstanceInput) (*rds.ModifyDBInstanceOutput, error) {
	fake.modifyDBInstanceMutex.Lock()
	ret, specificReturn := fake.modifyDBInstanceReturnsOnCall[len(fake.modifyDBInstanceArgsForCall)]
	fake.modifyDBInstanceArgsForCall = append(fake.modifyDBInstanceArgsForCall, struct {
		input *rds.ModifyDBInstanceInput
	}{input})
	fake.recordInvocation("ModifyDBInstance", []interface{}{input})
	fake.modifyDBInstanceMutex.Unlock()
	if fake.ModifyDBInstanceStub != nil {
		return fake.ModifyDBInstanceStub(input)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.modifyDBInstanceReturns.result1, fake.modifyDBInstanceReturns.result2
}

func (fake *FakeRDSService) ModifyDBInstanceCallCount() int {
	fake.modifyDBInstanceMutex.RLock()
	defer fake.modifyDBInstanceMutex.RUnlock()
	return len(fake.modifyDBInstanceArgsForCall)
}

func (fake *FakeRDSService) ModifyDBInstanceArgsForCall(i int) *rds.ModifyDBInstanceInput {
	fake.modifyDBInstanceMutex.RLock()
	defer fake.modifyDBInstanceMutex.RUnlock()
	return fake.modifyDBInstanceArgsForCall[i].input
}

func (fake *FakeRDSService) ModifyDBInstanceReturns(result1 *rds.ModifyDBInstanceOutput, result2 error) {
	fake.ModifyDBInstanceStub = nil
	fake.modifyDBInstanceReturns = struct {
		result1 *rds.ModifyDBInstanceOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) ModifyDBInstanceReturnsOnCall(i int, result1 *rds.ModifyDBInstanceOutput, result2 error) {
	fake.ModifyDBInstanceStub = nil
	if fake.modifyDBInstanceReturnsOnCall == nil {
		fake.modifyDBInstanceReturnsOnCall = make(map[int]struct {
			result1 *rds.ModifyDBInstanceOutput
			result2 error
		})
	}
	fake.modifyDBInstanceReturnsOnCall[i] = struct {
		result1 *rds.ModifyDBInstanceOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeRDSService) WaitUntilDBInstanceAvailable(input *rds.DescribeDBInstancesInput) error {
	fake.waitUntilDBInstanceAvailableMutex.Lock()
	ret, specificReturn := fake.waitUntilDBInstanceAvailableReturnsOnCall[len(fake.waitUntilDBInstanceAvailableArgsForCall)]
	fake.waitUntilDBInstanceAvailableArgsForCall = append(fake.waitUntilDBInstanceAvailableArgsForCall, struct {
		input *rds.DescribeDBInstancesInput
	}{input})
	fake.recordInvocation("WaitUntilDBInstanceAvailable", []interface{}{input})
	fake.waitUntilDBInstanceAvailableMutex.Unlock()
	if fake.WaitUntilDBInstanceAvailableStub != nil {
		return fake.WaitUntilDBInstanceAvailableStub(input)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.waitUntilDBInstanceAvailableReturns.result1
}

func (fake *FakeRDSService) WaitUntilDBInstanceAvailableCallCount() int {
	fake.waitUntilDBInstanceAvailableMutex.RLock()
	defer fake.waitUntilDBInstanceAvailableMutex.RUnlock()
	return len(fake.waitUntilDBInstanceAvailableArgsForCall)
}

func (fake *FakeRDSService) WaitUntilDBInstanceAvailableArgsForCall(i int) *rds.DescribeDBInstancesInput {
	fake.waitUntilDBInstanceAvailableMutex.RLock()
	defer fake.waitUntilDBInstanceAvailableMutex.RUnlock()
	return fake.waitUntilDBInstanceAvailableArgsForCall[i].input
}

func (fake *FakeRDSService) WaitUntilDBInstanceAvailableReturns(result1 error) {
	fake.WaitUntilDBInstanceAvailableStub = nil
	fake.waitUntilDBInstanceAvailableReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRDSService) WaitUntilDBInstanceAvailableReturnsOnCall(i int, result1 error) {
	fake.WaitUntilDBInstanceAvailableStub = nil
	if fake.waitUntilDBInstanceAvailableReturnsOnCall == nil {
		fake.waitUntilDBInstanceAvailableReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.waitUntilDBInstanceAvailableReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRDSService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.describeDBSubnetGroupsMutex.RLock()
	defer fake.describeDBSubnetGroupsMutex.RUnlock()
	fake.createDBInstanceMutex.RLock()
	defer fake.createDBInstanceMutex.RUnlock()
	fake.describeDBInstancesMutex.RLock()
	defer fake.describeDBInstancesMutex.RUnlock()
	fake.modifyDBInstanceMutex.RLock()
	defer fake.modifyDBInstanceMutex.RUnlock()
	fake.waitUntilDBInstanceAvailableMutex.RLock()
	defer fake.waitUntilDBInstanceAvailableMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRDSService) recordInvocation(key string, args []interface{}) {
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

var _ api.RDSService = new(FakeRDSService)
