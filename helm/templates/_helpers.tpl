{{- define "imagePullSecret" }}
{{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .Values.image.registry (printf "%s:%s" .Values.image.username .Values.image.password | b64enc) | b64enc }}
{{- end }}

{{- define "imageFull" -}}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.name .Values.image.tag -}}
{{- end -}}
