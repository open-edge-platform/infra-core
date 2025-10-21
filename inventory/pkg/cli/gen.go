// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

//go:generate go run gen.go

package main

import (
	"os"
	"strings"
	"text/template"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	"google.golang.org/protobuf/proto"
)

const (
	licenseHeader = "Apache" + "-2.0" // Break up to avoid license checker
)

var base = `
import (
	"context"
	"errors"
	//"flag"
	"fmt"
	//"os"
	//"os/signal"
	//"strings"
	//"sync"
	//"syscall"

	"github.com/manifoldco/promptui"
	"google.golang.org/protobuf/encoding/prototext"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_client "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	inv_errors "github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_util "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)
`

var tmplRaw = `
// Helper funcs for {{.ShortName}}
func helperList{{.ShortName}}s(ctx context.Context, client inv_client.InventoryClient) ([]*{{.ProtoPackageName}}.{{.FullMessageName}}, error) {
	kind, err := inv_util.GetResourceKindFromMessage(&{{.ProtoPackageName}}.{{.FullMessageName}}{})
	if err != nil {
		return nil, err
	}
	res, err := inv_util.GetResourceFromKind(kind)
	if err != nil {
		return nil, err
	}
	filter := &inv_v1.ResourceFilter{
		Resource:  res,
	}
	resp, err := client.ListAll(ctx, filter)
	if inv_errors.IsNotFound(err) {
		// Continue with empty list.
	} else if err != nil {
		return nil, err
	}
	return inv_util.GetSpecificResourceList[*{{.ProtoPackageName}}.{{.FullMessageName}}](resp)
}

func (c *Cli) PromptList{{.ShortName}}s(interface{}) (interface{}, error) {
	for {
		res, err := helperList{{.ShortName}}s(c.ctx, c.client)
		if err != nil {
			return nil, err
		}
		items := []menuItem{
			{Name: labelBack, Next: parentPrompt},
		}
		for _, r := range res {
			items = append(items, menuItem{
				Name: fmt.Sprintf("%.*s", c.lineItemMaxWidth, r),
				Next: c.Prompt{{.ShortName}}Details,
				Arg:  r,
			})
		}
		prompt := promptui.Select{
			Label:     fmt.Sprintf("%d {{.ShortName}}s total:", len(res)),
			Items:     items,
			Templates: selectTemplate,
			Size:      6,
			Searcher:  stringContainSearcher(items),
		}
		i, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return nil, err
		}
		arg, err := items[i].Next(items[i].Arg)
		if errors.Is(err, errPromptDone) {
			return arg, nil
		} else if errors.Is(err, errPromptSelectDone) {
			return arg, err
		} else if err != nil {
			return nil, err
		}
	}
}

func (c *Cli) Prompt{{.ShortName}}Details(arg interface{}) (interface{}, error) {
	for {
		r, ok := arg.(*{{.ProtoPackageName}}.{{.FullMessageName}})
		if !ok {
			return nil, errors.New("not a {{.ShortName}}")
		}
		fmt.Printf("%s", prototext.MarshalOptions{Multiline: true}.Format(r))
		items := []menuItem{
			{Name: labelBack, Next: parentPrompt, Arg: r},
			{Name: "<Select>", Next: returnSelectedItem, Arg: r},
			{Name: "<Delete>", Next: c.PromptDelete{{.ShortName}}, Arg: r},
		}
		prompt := promptui.Select{
			Label:     "{{.ShortName}} Actions:",
			Items:     items,
			Templates: selectTemplate,
			Size:      15,
		}
		i, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return nil, err
		}
		arg, err := items[i].Next(items[i].Arg)
		if errors.Is(err, errPromptDone) {
			return arg, nil
		} else if errors.Is(err, errPromptSelectDone) {
			return arg, err
		} else if err != nil {
			return nil, err
		}
	}
}

func (c *Cli) PromptDelete{{.ShortName}}(arg interface{}) (interface{}, error) {
	r, ok := arg.(*{{.ProtoPackageName}}.{{.FullMessageName}})
	if !ok {
		return nil, errors.New("not a {{.ShortName}}")
	}
	prompt := promptui.Prompt{
		Label:     "Confirm {{.ShortName}} deletion:",
		IsConfirm: true,
	}
	for {
		result, err := prompt.Run()
		if err != nil {
			return nil, nil
		}
		if result != "y" {
			return nil, nil
		}
		_, err = c.client.Delete(c.ctx, r.GetResourceId())
		return nil, err
	}
}
`

var tmpl = template.Must(template.New("foo").Parse(tmplRaw))

type resource struct {
	msg              proto.Message // Protobuf message defining the resource.
	ProtoPackageName string        // Name of the package defining this message, as imported in Go. E.g. "computev1".
	ShortName        string        // Human-readable short name of the resource. E.g. "VM".
	FullMessageName  string        // Auto-set.
}

func main() {
	out := strings.Builder{}
	out.WriteString(`// SPDX-FileCopyrightText: (C) 2025 Intel Corporation`)
	out.WriteString(`// SPDX-License-Identifier: ` + licenseHeader)
	out.WriteString("\n")
	out.WriteString(`// Code generated by "gen", DO NOT EDIT.`)
	out.WriteString("\n\n")
	out.WriteString("package cli\n\n")
	out.WriteString(base)

	resources := []resource{
		{msg: &computev1.HostResource{}, ShortName: "Host", ProtoPackageName: "computev1"},
		{msg: &computev1.HostnicResource{}, ShortName: "HostNic", ProtoPackageName: "computev1"},
		{msg: &computev1.HoststorageResource{}, ShortName: "HostStorage", ProtoPackageName: "computev1"},
		{msg: &computev1.HostusbResource{}, ShortName: "HostUSB", ProtoPackageName: "computev1"},
		{msg: &computev1.InstanceResource{}, ShortName: "Instance", ProtoPackageName: "computev1"},
		{msg: &computev1.WorkloadResource{}, ShortName: "Workload", ProtoPackageName: "computev1"},
		{msg: &computev1.WorkloadMember{}, ShortName: "WorkloadMember", ProtoPackageName: "computev1"},

		{msg: &locationv1.RegionResource{}, ShortName: "Region", ProtoPackageName: "locationv1"},
		{msg: &locationv1.SiteResource{}, ShortName: "Site", ProtoPackageName: "locationv1"},

		// TODO(max): add all other resources
	}

	for _, r := range resources {
		r.FullMessageName = string(r.msg.ProtoReflect().Descriptor().Name())
		if err := tmpl.Execute(&out, r); err != nil {
			panic(err)
		}
	}

	if err := os.WriteFile("cli_generated.go", []byte(out.String()), 0644); err != nil {
		panic(err)
	}
}
