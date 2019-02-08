FORMAT: 1A

# Developers Platform Service
This is description of DPS project.

## Actions
{{- range . }}
  {{- if .Usage }}
+ [{{ .Usage }}](#endpoint-{{ .ID }})
  {{- end }}
{{- end }}

{{- range . }}
{{- if or .Usage }}
{{- $requestHeaders := index (.Definition.Objects "header") "header" }}
{{- $requestBody := .Definition.Request.JSON 10 }}
{{- if eq "null" $requestBody }}
    {{- $requestBody := "" }}      
{{- end }}
 
{{- $responseBody := .Definition.Response.JSON 10 }}
{{- $parameters := index (.Definition.Objects "parameter") "parameter" }}
          
# <a name="endpoint-{{ .ID }}"></a>{{ .Usage }} [{{ .Method }} {{ .Path }}]
{{- if .Description }}
    {{ "\r" }}{{ .Description }}{{ "\n" }}
{{- end}}

{{- if or .Definition.Request .Definition.Headers }}
+ Request (application/json)
  {{- if .Definition.Parameters }}
    + Parameters
    {{""}} 
    {{- range $k, $field := $parameters.Vars }}
      {{- if not $field.Ignored }}
        {{- $action := "optional" }}
        {{- if $field.Required }}
          {{- $action = "required" }}
        {{- end }}
        + {{ $field.NameR "parameter" }}: {{ $field.Example }} ({{ $field.Type }}, {{ $action }}) - {{ $field.Description }}
      {{- end }} 
    {{- end}}
    {{""}} 
  {{- end }}
  {{- if not .Definition.Headers.Root.Ignored  }}
    + Headers
     {{- range $k, $field := $requestHeaders.Vars }}
         {{- if not $field.Ignored }}             
             {{ $field.NameR "header" }}: {{ $field.Example }}
         {{- end }}
     {{- end }}
  {{- end }} 
  {{- if .Definition.Request }}
    + Attributes ({{ .Definition.Request.Root.Type }})
    {{- range $k, $field := .Definition.Request.ByName }}
      {{- if not $field.IsRoot }}
      {{- if not $field.Ignored }}
        {{- $action := "optional" }}
          {{- if $field.Required }}
            {{- $action = "required" }}
          {{- end }}
        {{- if $field.Object }}
        elo
        {{- end }}  
        + {{ $field.NormalisedPath }} ({{ $field.Type }}, {{ $action }}) - {{ $field.Description }}
      {{- end }}
      {{- end }} 
    {{- end}}
    {{ "" }}
  {{- end}}  
  {{- if ne $requestBody "null" }}
    + Body
    
            {{ .Definition.Request.JSON 12 }}
        
  {{- end }}
{{- end }}

{{- if .Definition.Response }}

+ Response 200 (application/json)

    + Attributes ({{ .Definition.Response.Root.Type }})
    {{- range $k, $field := .Definition.Response.ByName }}
      {{- if not $field.Ignored }}
      {{- if not $field.IsRoot }}
        + {{ $field.NormalisedPath }} ({{ $field.Type }}) - {{ $field.Description }}
      {{- end }}  
      {{- end }} 
    {{- end}}

    + Body  
          
            {{ .Definition.Response.JSON 12 }}
            
{{- end }}
{{- end }}
{{- end }}