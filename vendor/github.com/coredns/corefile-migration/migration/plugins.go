package migration

import (
	"errors"

	"github.com/coredns/corefile-migration/migration/corefile"
)

type plugin struct {
	status         string
	replacedBy     string
	additional     string
	namedOptions   map[string]option
	patternOptions map[string]option
	action         pluginActionFn // action affecting this plugin only
	add            serverActionFn // action to add a new plugin to the server block
	downAction     pluginActionFn // downgrade action affecting this plugin only
}

type option struct {
	name       string
	status     string
	replacedBy string
	additional string
	action     optionActionFn // action affecting this option only
	add        pluginActionFn // action to add the option to the plugin
	downAction optionActionFn // downgrade action affecting this option only
}

type corefileAction func(*corefile.Corefile) (*corefile.Corefile, error)
type serverActionFn func(*corefile.Server) (*corefile.Server, error)
type pluginActionFn func(*corefile.Plugin) (*corefile.Plugin, error)
type optionActionFn func(*corefile.Option) (*corefile.Option, error)

// plugins holds a map of plugin names and their migration rules per "version".  "Version" here is meaningless outside
// of the context of this code. Each change in options or migration actions for a plugin requires a new "version"
// containing those new/removed options and migration actions. Plugins in CoreDNS are not versioned.
var plugins = map[string]map[string]plugin{
	"kubernetes": {
		"v1": plugin{
			namedOptions: map[string]option{
				"resyncperiod":       {},
				"endpoint":           {},
				"tls":                {},
				"namespaces":         {},
				"labels":             {},
				"pods":               {},
				"endpoint_pod_names": {},
				"upstream":           {},
				"ttl":                {},
				"noendpoints":        {},
				"transfer":           {},
				"fallthrough":        {},
				"ignore":             {},
			},
		},
		"v2": plugin{
			namedOptions: map[string]option{
				"resyncperiod":       {},
				"endpoint":           {},
				"tls":                {},
				"namespaces":         {},
				"labels":             {},
				"pods":               {},
				"endpoint_pod_names": {},
				"upstream":           {},
				"ttl":                {},
				"noendpoints":        {},
				"transfer":           {},
				"fallthrough":        {},
				"ignore":             {},
				"kubeconfig":         {}, // new option
			},
		},
		"v3": plugin{
			namedOptions: map[string]option{
				"resyncperiod": {},
				"endpoint": { // new deprecation
					status: deprecated,
					action: useFirstArgumentOnly,
				},
				"tls":                {},
				"kubeconfig":         {},
				"namespaces":         {},
				"labels":             {},
				"pods":               {},
				"endpoint_pod_names": {},
				"upstream":           {},
				"ttl":                {},
				"noendpoints":        {},
				"transfer":           {},
				"fallthrough":        {},
				"ignore":             {},
			},
		},
		"v4": plugin{
			namedOptions: map[string]option{
				"resyncperiod": {},
				"endpoint": {
					status: ignored,
					action: useFirstArgumentOnly,
				},
				"tls":                {},
				"kubeconfig":         {},
				"namespaces":         {},
				"labels":             {},
				"pods":               {},
				"endpoint_pod_names": {},
				"upstream": { // new deprecation
					status: deprecated,
					action: removeOption,
				},
				"ttl":         {},
				"noendpoints": {},
				"transfer":    {},
				"fallthrough": {},
				"ignore":      {},
			},
		},
		"v5": plugin{
			namedOptions: map[string]option{
				"resyncperiod": { // new deprecation
					status: deprecated,
					action: removeOption,
				},
				"endpoint": {
					status: ignored,
					action: useFirstArgumentOnly,
				},
				"tls":                {},
				"kubeconfig":         {},
				"namespaces":         {},
				"labels":             {},
				"pods":               {},
				"endpoint_pod_names": {},
				"upstream": {
					status: ignored,
					action: removeOption,
				},
				"ttl":         {},
				"noendpoints": {},
				"transfer":    {},
				"fallthrough": {},
				"ignore":      {},
			},
		},
		"v6": plugin{
			namedOptions: map[string]option{
				"resyncperiod": { // now ignored
					status: ignored,
					action: removeOption,
				},
				"endpoint": {
					status: ignored,
					action: useFirstArgumentOnly,
				},
				"tls":                {},
				"kubeconfig":         {},
				"namespaces":         {},
				"labels":             {},
				"pods":               {},
				"endpoint_pod_names": {},
				"upstream": {
					status: ignored,
					action: removeOption,
				},
				"ttl":         {},
				"noendpoints": {},
				"transfer":    {},
				"fallthrough": {},
				"ignore":      {},
			},
		},
		"v7": plugin{
			namedOptions: map[string]option{
				"resyncperiod": { // new removal
					status: removed,
					action: removeOption,
				},
				"endpoint": {
					status: ignored,
					action: useFirstArgumentOnly,
				},
				"tls":                {},
				"kubeconfig":         {},
				"namespaces":         {},
				"labels":             {},
				"pods":               {},
				"endpoint_pod_names": {},
				"upstream": { // new removal
					status: removed,
					action: removeOption,
				},
				"ttl":         {},
				"noendpoints": {},
				"transfer":    {},
				"fallthrough": {},
				"ignore":      {},
			},
		},
	},

	"errors": {
		"v1": plugin{},
		"v2": plugin{
			namedOptions: map[string]option{
				"consolidate": {},
			},
		},
	},

	"health": {
		"v1": plugin{
			namedOptions: map[string]option{
				"lameduck": {},
			},
		},
		"v1 add lameduck": plugin{
			namedOptions: map[string]option{
				"lameduck": {
					status: newdefault,
					add: func(c *corefile.Plugin) (*corefile.Plugin, error) {
						return addOptionToPlugin(c, &corefile.Option{Name: "lameduck 5s"})
					},
					downAction: removeOption,
				},
			},
		},
	},

	"hosts": {
		"v1": plugin{
			namedOptions: map[string]option{
				"ttl":         {},
				"no_reverse":  {},
				"reload":      {},
				"fallthrough": {},
			},
			patternOptions: map[string]option{
				`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`:              {}, // close enough
				`[0-9A-Fa-f]{1,4}:[:0-9A-Fa-f]+:[0-9A-Fa-f]{1,4}`: {}, // less close, but still close enough
			},
		},
	},

	"rewrite": {
		"v1": plugin{
			namedOptions: map[string]option{
				"type":        {},
				"class":       {},
				"name":        {},
				"answer name": {},
				"edns0":       {},
			},
		},
		"v2": plugin{
			namedOptions: map[string]option{
				"type":        {},
				"class":       {},
				"name":        {},
				"answer name": {},
				"edns0":       {},
				"ttl":         {}, // new option
			},
		},
	},

	"log": {
		"v1": plugin{
			namedOptions: map[string]option{
				"class": {},
			},
		},
	},

	"cache": {
		"v1": plugin{
			namedOptions: map[string]option{
				"success":  {},
				"denial":   {},
				"prefetch": {},
			},
		},
		"v2": plugin{
			namedOptions: map[string]option{
				"success":     {},
				"denial":      {},
				"prefetch":    {},
				"serve_stale": {}, // new option
			},
		},
	},

	"forward": {
		"v1": plugin{
			namedOptions: map[string]option{
				"except":         {},
				"force_tcp":      {},
				"expire":         {},
				"max_fails":      {},
				"tls":            {},
				"tls_servername": {},
				"policy":         {},
				"health_check":   {},
			},
		},
		"v2": plugin{
			namedOptions: map[string]option{
				"except":         {},
				"force_tcp":      {},
				"prefer_udp":     {},
				"expire":         {},
				"max_fails":      {},
				"tls":            {},
				"tls_servername": {},
				"policy":         {},
				"health_check":   {},
			},
		},
		"v3": plugin{
			namedOptions: map[string]option{
				"except":         {},
				"force_tcp":      {},
				"prefer_udp":     {},
				"expire":         {},
				"max_fails":      {},
				"tls":            {},
				"tls_servername": {},
				"policy":         {},
				"health_check":   {},
				"max_concurrent": { // new option
					status: newdefault,
					add: func(c *corefile.Plugin) (*corefile.Plugin, error) {
						return addOptionToPlugin(c, &corefile.Option{Name: "max_concurrent 1000"})
					},
					downAction: removeOption,
				},
			},
		},
	},

	"k8s_external": {
		"v1": plugin{
			namedOptions: map[string]option{
				"apex": {},
				"ttl":  {},
			},
		},
	},

	"proxy": {
		"v1": plugin{
			namedOptions: map[string]option{
				"policy":       {},
				"fail_timeout": {},
				"max_fails":    {},
				"health_check": {},
				"except":       {},
				"spray":        {},
				"protocol": { // https_google option ignored
					status: ignored,
					action: proxyRemoveHttpsGoogleProtocol,
				},
			},
		},
		"v2": plugin{
			namedOptions: map[string]option{
				"policy":       {},
				"fail_timeout": {},
				"max_fails":    {},
				"health_check": {},
				"except":       {},
				"spray":        {},
				"protocol": { // https_google option removed
					status: removed,
					action: proxyRemoveHttpsGoogleProtocol,
				},
			},
		},
		"deprecation": plugin{ // proxy -> forward deprecation migration
			status:       deprecated,
			replacedBy:   "forward",
			action:       proxyToForwardPluginAction,
			namedOptions: proxyToForwardOptionsMigrations,
		},
		"removal": plugin{ // proxy -> forward forced migration
			status:       removed,
			replacedBy:   "forward",
			action:       proxyToForwardPluginAction,
			namedOptions: proxyToForwardOptionsMigrations,
		},
	},
}

func removePlugin(*corefile.Plugin) (*corefile.Plugin, error) { return nil, nil }
func removeOption(*corefile.Option) (*corefile.Option, error) { return nil, nil }

func renamePlugin(p *corefile.Plugin, to string) (*corefile.Plugin, error) {
	p.Name = to
	return p, nil
}

func addToServerBlockWithPlugins(sb *corefile.Server, newPlugin *corefile.Plugin, with []string) (*corefile.Server, error) {
	if len(with) == 0 {
		// add to all blocks
		sb.Plugins = append(sb.Plugins, newPlugin)
		return sb, nil
	}
	for _, p := range sb.Plugins {
		for _, w := range with {
			if w == p.Name {
				// add to this block
				sb.Plugins = append(sb.Plugins, newPlugin)
				return sb, nil
			}
		}
	}
	return sb, nil
}

func addToKubernetesServerBlocks(sb *corefile.Server, newPlugin *corefile.Plugin) (*corefile.Server, error) {
	return addToServerBlockWithPlugins(sb, newPlugin, []string{"kubernetes"})
}

func addToForwardingServerBlocks(sb *corefile.Server, newPlugin *corefile.Plugin) (*corefile.Server, error) {
	return addToServerBlockWithPlugins(sb, newPlugin, []string{"forward", "proxy"})
}

func addToAllServerBlocks(sb *corefile.Server, newPlugin *corefile.Plugin) (*corefile.Server, error) {
	return addToServerBlockWithPlugins(sb, newPlugin, []string{})
}

func addOptionToPlugin(pl *corefile.Plugin, newOption *corefile.Option) (*corefile.Plugin, error) {
	pl.Options = append(pl.Options, newOption)
	return pl, nil
}

var proxyToForwardOptionsMigrations = map[string]option{
	"policy": {
		action: func(o *corefile.Option) (*corefile.Option, error) {
			if len(o.Args) == 1 && o.Args[0] == "least_conn" {
				o.Name = "force_tcp"
				o.Args = nil
			}
			return o, nil
		},
	},
	"except":       {},
	"fail_timeout": {action: removeOption},
	"max_fails":    {action: removeOption},
	"health_check": {action: removeOption},
	"spray":        {action: removeOption},
	"protocol": {
		action: func(o *corefile.Option) (*corefile.Option, error) {
			if len(o.Args) >= 2 && o.Args[0] == "force_tcp" {
				o.Name = "force_tcp"
				o.Args = nil
				return o, nil
			}
			return nil, nil
		},
	},
}

var proxyToForwardPluginAction = func(p *corefile.Plugin) (*corefile.Plugin, error) {
	return renamePlugin(p, "forward")
}

var useFirstArgumentOnly = func(o *corefile.Option) (*corefile.Option, error) {
	if len(o.Args) < 1 {
		return o, nil
	}
	o.Args = o.Args[:1]
	return o, nil
}

var proxyRemoveHttpsGoogleProtocol = func(o *corefile.Option) (*corefile.Option, error) {
	if len(o.Args) > 0 && o.Args[0] == "https_google" {
		return nil, nil
	}
	return o, nil
}

func breakForwardStubDomainsIntoServerBlocks(cf *corefile.Corefile) (*corefile.Corefile, error) {
	for _, sb := range cf.Servers {
		for j, fwd := range sb.Plugins {
			if fwd.Name != "forward" {
				continue
			}
			if len(fwd.Args) == 0 {
				return nil, errors.New("found invalid forward plugin declaration")
			}
			if fwd.Args[0] == "." {
				// dont move the default upstream
				continue
			}
			if len(sb.DomPorts) != 1 {
				return cf, errors.New("unhandled migration of multi-domain/port server block")
			}
			if sb.DomPorts[0] != "." && sb.DomPorts[0] != ".:53" {
				return cf, errors.New("unhandled migration of non-default domain/port server block")
			}

			newSb := &corefile.Server{}                // create a new server block
			newSb.DomPorts = []string{fwd.Args[0]}     // copy the forward zone to the server block domain
			fwd.Args[0] = "."                          // the plugin's zone changes to "." for brevity
			newSb.Plugins = append(newSb.Plugins, fwd) // add the plugin to its new server block

			// Add appropriate addtl plugins to new server block
			newSb.Plugins = append(newSb.Plugins, &corefile.Plugin{Name: "loop"})
			newSb.Plugins = append(newSb.Plugins, &corefile.Plugin{Name: "errors"})
			newSb.Plugins = append(newSb.Plugins, &corefile.Plugin{Name: "cache", Args: []string{"30"}})

			//add new server block to corefile
			cf.Servers = append(cf.Servers, newSb)

			//remove the forward plugin from the original server block
			sb.Plugins = append(sb.Plugins[:j], sb.Plugins[j+1:]...)
		}
	}
	return cf, nil
}
