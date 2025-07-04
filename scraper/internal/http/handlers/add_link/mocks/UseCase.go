// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	scraper "scraper/internal/model/scraper"
	mock "github.com/stretchr/testify/mock"
)

// UseCase is an autogenerated mock type for the UseCase type
type UseCase struct {
	mock.Mock
}

type UseCase_Expecter struct {
	mock *mock.Mock
}

func (_m *UseCase) EXPECT() *UseCase_Expecter {
	return &UseCase_Expecter{mock: &_m.Mock}
}

// AddLink provides a mock function with given fields: ctx, id, link
func (_m *UseCase) AddLink(ctx context.Context, id int64, link *scraper.Link) (scraper.Link, error) {
	ret := _m.Called(ctx, id, link)

	if len(ret) == 0 {
		panic("no return value specified for AddLink")
	}

	var r0 scraper.Link
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *scraper.Link) (scraper.Link, error)); ok {
		return rf(ctx, id, link)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *scraper.Link) scraper.Link); ok {
		r0 = rf(ctx, id, link)
	} else {
		r0 = ret.Get(0).(scraper.Link)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *scraper.Link) error); ok {
		r1 = rf(ctx, id, link)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UseCase_AddLink_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddLink'
type UseCase_AddLink_Call struct {
	*mock.Call
}

// AddLink is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
//   - link *scraper.Link
func (_e *UseCase_Expecter) AddLink(ctx interface{}, id interface{}, link interface{}) *UseCase_AddLink_Call {
	return &UseCase_AddLink_Call{Call: _e.mock.On("AddLink", ctx, id, link)}
}

func (_c *UseCase_AddLink_Call) Run(run func(ctx context.Context, id int64, link *scraper.Link)) *UseCase_AddLink_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(*scraper.Link))
	})
	return _c
}

func (_c *UseCase_AddLink_Call) Return(_a0 scraper.Link, _a1 error) *UseCase_AddLink_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *UseCase_AddLink_Call) RunAndReturn(run func(context.Context, int64, *scraper.Link) (scraper.Link, error)) *UseCase_AddLink_Call {
	_c.Call.Return(run)
	return _c
}

// NewUseCase creates a new instance of UseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *UseCase {
	mock := &UseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
