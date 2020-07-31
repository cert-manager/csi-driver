/*
Copyright 2019 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templating

import (
	"bytes"
	"text/template"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func RenderAttributeTemplates(attr map[string]string) (map[string]string, error) {
	var err error

	tmpls := &template.Template{}
	if err = addAndParseTemplate(tmpls, csiapi.CommonNameKey); err != nil {
		return nil, err
	}
	if err = addAndParseTemplate(tmpls, csiapi.DNSNamesKey); err != nil {
		return nil, err
	}
	if err = addAndParseTemplate(tmpls, csiapi.URISANsKey); err != nil {
		return nil, err
	}

	tmplData := csiapi.TemplatingData{
		PodName:      attr[csiapi.CSIPodNameKey],
		PodNamespace: attr[csiapi.CSIPodNamespaceKey],
		PodUID:       attr[csiapi.CSIPodUIDKey],
	}

	for _, t := range tmpls.Templates() {
		buf := new(bytes.Buffer)
		if err := t.Execute(buf, tmplData); err != nil {
			return nil, err
		}
		attr[t.Name()] = buf.String()
	}

	return attr, nil
}

func addAndParseTemplate(t *template.Template, k string) error {
	if _, err := t.New(k).Parse(k); err != nil {
		return err
	}
	return nil
}
