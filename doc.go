// package dynjson allow APIs to return only fields selected by the API client:
//
// 	GET https://api.example.com/v1/foos
// 	[{"id":1,foo":1,"bar":2,"baz":3}]
//
// 	GET https://api.example.com/v1/foos?select=foo
// 	[{"foo":1}]
//
// 	GET https://api.example.com/v1/foos/1?select=foo
// 	{"foo":1}
//
// dynjson mimicks the original struct using the original types and json tags.
// The field order is the same as the select parameters.
package dynjson
