package ziface

/*
路由里的数据都是IRequest
*/

type IRouter interface {
	// 在处理conn业务之前的钩子方法
	PreHandle(request IRequest)

	// 在处理conn业务的主方法
	Handle(request IRequest)

	// 在处理conn业务之后的钩子方法
	PostHandle(request IRequest)

}