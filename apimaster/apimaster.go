package apimaster

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"reflect"

	"github.com/rogue-syntax/rs-goapiserver/apireturn"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
)

var builtInTypes = []string{
	"bool",
	"byte",
	"complex64", "complex128",
	"error",
	"float32", "float64",
	"int", "int8", "int16", "int32", "int64",
	"rune",
	"string",
	"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
}

type StructDescriptorMap map[string]map[string]interface{}

func MakeStructDescriptorJSON(s interface{}) string {
	t := reflect.TypeOf(&s)

	fields := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields[field.Name] = field.Type.String()
	}

	jsonData, _ := json.Marshal(fields)
	return string(jsonData)
}

type ApiNilDescriptor struct{}

// generates a map of the fields of a struct and their types
// Uee like : MakeStructDescriptorMap(new(ExampleInput))
// Use an ApiNilDescriptor if there is no input like : MakeStructDescriptorMap(new(ApiNilDescriptor))
func MakeStructDescriptorMap[T any](s *T) StructDescriptorMap {

	fields := make(map[string]interface{})
	nameMap := make(map[string]map[string]interface{})

	t := reflect.TypeOf(*s)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields[field.Name] = field.Type.String()
	}
	nameMap[t.Name()] = fields
	return nameMap
}

type TypeNameString string
type RouteParamSource string

const (
	URIREQ         RouteParamSource = "URI"
	GETREQ         RouteParamSource = "GET"
	POSTREQ        RouteParamSource = "POST"
	MULTIPART_FORM RouteParamSource = "MULTIPART-FROM"
)

type RouteParam struct {
	Source RouteParamSource
}
type ApiReqDef struct {
	API           string
	Method        RouteParamSource
	Desc          string
	Input         StructDescriptorMap
	OutputData    StructDescriptorMap
	OutputWrapper StructDescriptorMap
}

type ExampleInput struct {
	ExampleString string
	ExampleInt    int
}
type ExampleOutput struct {
	ExampleString string
	ExampleInt    int
}

var ApiReqMap map[string]map[string]ApiReqDef = map[string]map[string]ApiReqDef{
	"examples": {
		"/v1/example": {
			API:           "/v1/example",
			Method:        GETREQ,
			Desc:          "Example API",
			Input:         MakeStructDescriptorMap(new(ExampleInput)),
			OutputData:    MakeStructDescriptorMap(new(ExampleOutput)),
			OutputWrapper: MakeStructDescriptorMap(new(apireturn.JsonReturn)),
		},
	},
}

func Handler_GetApiReqMap(w http.ResponseWriter, r *http.Request, ctx context.Context) {

	apireturn.ApiJSONReturn(ApiReqMap, apierrorkeys.NOError, &w)

}

func Handler_GetApiReqMapPage(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	templateStr := `
	<h1>API</h1>
		<ul>
		{{ range $key, $value := .}}
			<li>
				<div> 
					{{$key}}
				</div>
				<div>
					<ul>
					{{ range $key2, $value2 := $value}}
						<li>
							<div>
								{{$key2}}
							</div>
							<div>
								<table>
								
									<tr>	
										<td>
											Route:
										</td>
										<td>
											{{$value2.API}}
										</td>
									</tr>
									<tr>	
										<td>
											Request Method:
										</td>
										<td>
											{{$value2.Method}}
										</td>
									</tr>
									<tr>	
										<td>
											Description:
										</td>
										<td>
											{{$value2.Desc}}
										</td>
									</tr>
									<tr>
										<td>
											Input:
										</td>
										<td>
										{{ range $key3, $value3 := $value2.Input}}
											<div>
												{{$key3}}
											</div>
											<div>
												<table>
												{{ range $key4, $value4 := $value3}}
													<tr>
														<td>
															{{$key4}}
														</td>
														<td>
															{{$value4}}
														</td>
													</tr>
												{{ end }}
												</table>
											</div>
										{{ end }}	
										</td>
									</tr>
									<tr>
										<td>
											Output Data:
										</td>
										<td>
											<table>
											{{ range $key3, $value3 := $value2.OutputData}}

												<tr>
													<td>

														{{$key3}}	
													</td>
													<td>
														{{$value3}}
													</td>
												</tr>

											{{ end }}
											</table>
										</td>
									</tr>
									<tr>
										<td>
											Output Wrapper:
										</td>
										<td>
											<table>
											{{ range $key3, $value3 := $value2.OutputWrapper}}

												<tr>
													<td>

														{{$key3}}	
													</td>
													<td>
														{{$value3}}
													</td>
												</tr>

											{{ end }}
											</table>
										</td>
									</tr>
								
								</table>
							</div>
						</li>
					{{ end }}
					</ul>
				</div>
			</li>
		{{ end }}
		</ul>
	`
	tmplt := template.Must(template.New("api-page").Parse(templateStr))
	tmplt.Execute(w, ApiReqMap)
	//apireturn.ApiJSONReturn(ApiReqMap, apierrorkeys.NOError, &w)

}
