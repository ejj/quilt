// Code generated by mockery v1.0.1 DO NOT EDIT.

package mocks

import api "github.com/hashicorp/vault/api"
import mock "github.com/stretchr/testify/mock"

// APIClient is an autogenerated mock type for the APIClient type
type APIClient struct {
	mock.Mock
}

// Delete provides a mock function with given fields: path
func (_m *APIClient) Delete(path string) (*api.Secret, error) {
	ret := _m.Called(path)

	var r0 *api.Secret
	if rf, ok := ret.Get(0).(func(string) *api.Secret); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeletePolicy provides a mock function with given fields: name
func (_m *APIClient) DeletePolicy(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnableAuth provides a mock function with given fields: path, authType, desc
func (_m *APIClient) EnableAuth(path string, authType string, desc string) error {
	ret := _m.Called(path, authType, desc)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(path, authType, desc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetPolicy provides a mock function with given fields: name
func (_m *APIClient) GetPolicy(name string) (string, error) {
	ret := _m.Called(name)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Init provides a mock function with given fields: opts
func (_m *APIClient) Init(opts *api.InitRequest) (*api.InitResponse, error) {
	ret := _m.Called(opts)

	var r0 *api.InitResponse
	if rf, ok := ret.Get(0).(func(*api.InitRequest) *api.InitResponse); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.InitResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*api.InitRequest) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InitStatus provides a mock function with given fields:
func (_m *APIClient) InitStatus() (bool, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: path
func (_m *APIClient) List(path string) (*api.Secret, error) {
	ret := _m.Called(path)

	var r0 *api.Secret
	if rf, ok := ret.Get(0).(func(string) *api.Secret); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPolicies provides a mock function with given fields:
func (_m *APIClient) ListPolicies() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutPolicy provides a mock function with given fields: name, rules
func (_m *APIClient) PutPolicy(name string, rules string) error {
	ret := _m.Called(name, rules)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(name, rules)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Read provides a mock function with given fields: path
func (_m *APIClient) Read(path string) (*api.Secret, error) {
	ret := _m.Called(path)

	var r0 *api.Secret
	if rf, ok := ret.Get(0).(func(string) *api.Secret); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetToken provides a mock function with given fields: token
func (_m *APIClient) SetToken(token string) {
	_m.Called(token)
}

// Unseal provides a mock function with given fields: shard
func (_m *APIClient) Unseal(shard string) (*api.SealStatusResponse, error) {
	ret := _m.Called(shard)

	var r0 *api.SealStatusResponse
	if rf, ok := ret.Get(0).(func(string) *api.SealStatusResponse); ok {
		r0 = rf(shard)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.SealStatusResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(shard)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Write provides a mock function with given fields: path, data
func (_m *APIClient) Write(path string, data map[string]interface{}) (*api.Secret, error) {
	ret := _m.Called(path, data)

	var r0 *api.Secret
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}) *api.Secret); ok {
		r0 = rf(path, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string]interface{}) error); ok {
		r1 = rf(path, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
