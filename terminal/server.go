package terminal

import (
	"github.com/fanux/fist/tools"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
)

//Register is
func Register(container *restful.Container) {
	terminal := new(restful.WebService)
	terminal.
		Path("/").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	terminal.Route(terminal.POST("/terminal").To(createTerminal))
	terminal.Route(terminal.GET("/heartbeat").To(handleHeartbeat))

	container.Add(terminal)
}

func createTerminal(request *restful.Request, response *restful.Response) {
	t := newTerminal()
	err := request.ReadEntity(t)
	if err != nil {
		tools.ResponseSystemError(response, err)
		return
	}

	err = t.Create()
	if err != nil {
		tools.ResponseSystemError(response, err)
		return
	}
	tools.ResponseSuccess(response, t)
}

func handleHeartbeat(request *restful.Request, response *restful.Response) {
	//get client of k8s
	clientset, err := tools.GetK8sClient()
	if err != nil {
		tools.ResponseSystemError(response, err)
		return
	}
	tid := request.QueryParameter("tid")
	if tid == "" {
		tools.ResponseSystemError(response, tools.ErrParamTidEmpty)
		return
	}
	namespace := request.QueryParameter("namespace")
	if namespace == "" {
		namespace = DefaultTTYnameapace
	}
	var hbInterface Heartbeater
	hbInterface = NewHeartbeater(tid, namespace)
	err = hbInterface.UpdateTimestamp(clientset)
	if err != nil {
		tools.ResponseSystemError(response, err)
		return
	}
	tools.ResponseSuccess(response, nil)
}

//Serve start a terminal server
func Serve() {
	LoadTerminalID()

	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	Register(wsContainer)

	log.Print("start listening on localhost:8080")
	server := &http.Server{Addr: ":8080", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}