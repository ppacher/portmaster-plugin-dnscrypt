package main

import (
	"context"
	"os"
	"sync"

	"github.com/ameshkov/dnscrypt/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/miekg/dns"
	"github.com/safing/portmaster/plugin/framework"
	"github.com/safing/portmaster/plugin/framework/cmds"
	"github.com/safing/portmaster/plugin/shared"
	"github.com/safing/portmaster/plugin/shared/proto"
	"github.com/spf13/cobra"
)

var (
	client dnscrypt.Client

	resolverLock sync.RWMutex
	resolverInfo *dnscrypt.ResolverInfo
)

func convertRRs(list []dns.RR) []*proto.DNSRR {
	var rrs []*proto.DNSRR
	for _, answer := range list {
		var (
			rType uint16
			rData []byte
		)

		switch v := answer.(type) {
		case *dns.A:
			rType = dns.TypeA
			rData = v.A
		case *dns.AAAA:
			rType = dns.TypeAAAA
			rData = v.AAAA
		case *dns.CNAME:
			rType = dns.TypeCNAME
			rData = []byte(v.Target)
		case *dns.TXT:
			rType = dns.TypeCNAME
			if len(v.Txt) > 0 {
				rData = []byte(v.Txt[0])
			}
		default:
			continue
		}

		rrs = append(rrs, &proto.DNSRR{
			Name:  answer.Header().Name,
			Type:  uint32(rType),
			Class: uint32(answer.Header().Class),
			Ttl:   answer.Header().Ttl,
			Data:  rData,
		})
	}

	return rrs
}

func resolve(ctx context.Context, question *proto.DNSQuestion, conn *proto.Connection) (*proto.DNSResponse, error) {
	resolverLock.RLock()
	defer resolverLock.RUnlock()

	if resolverInfo == nil {
		return nil, nil
	}

	req := &dns.Msg{}
	req.Id = dns.Id()
	req.RecursionDesired = true
	req.Question = []dns.Question{
		{
			Name:   question.Name,
			Qtype:  uint16(question.Type),
			Qclass: uint16(question.Class),
		},
	}

	result, err := client.Exchange(req, resolverInfo)
	if err != nil {
		return nil, err
	}

	// TODO(ppacher): add support for extra and NS as well.

	return &proto.DNSResponse{
		Rcode: uint32(result.Rcode),
		Rrs:   convertRRs(result.Answer),
	}, nil
}

func getResolverInfo(server string) {

	// Fetching and validating the server certificate
	info, err := client.Dial(server)
	if err != nil {
        _, err := framework.Notify().CreateNotification(framework.Context(), &proto.Notification{
			EventId: "dnscrypt-invalid-stamp",
			Title:   "DNSCrypt: Server Stamp invalid",
			Message: err.Error(),
		})
        if err != nil {
		    hclog.L().Error("failed to create notification", "error", err)
        }

		return
	}

	resolverLock.Lock()
	defer resolverLock.Unlock()

	resolverInfo = info
}

func setupAndWatchConfig(ctx context.Context) error {
	if err := framework.Config().RegisterOption(ctx, &proto.Option{
		Name:        "DNSCrypt Server",
		Description: "Stamp of the DNSCrypt server",
		Key:         "dnscryptServer",
		OptionType:  proto.OptionType_OPTION_TYPE_STRING,
		Default: &proto.Value{
			String_: "",
		},
	}); err != nil {
		return err
	}

	ch, err := framework.Config().WatchValue(framework.Context(), "dnscryptServer")
	if err != nil {
		return err
	}

	go func() {
		for msg := range ch {
			getResolverInfo(msg.Value.String_)
		}
	}()

	val, err := framework.Config().GetValue(ctx, "dnscryptServer")
	if err != nil {
		return err
	}

	if srv := val.String_; srv != "" {
		getResolverInfo(srv)
	}

	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use: "portmaster-plugin-dnscrypt",
		Run: func(cmd *cobra.Command, args []string) {
            err := framework.RegisterResolver(
				framework.ResolverFunc(resolve),
			)
            if err != nil {
                panic(err)
            }

			framework.OnInit(func(ctx context.Context) error {
				if err := setupAndWatchConfig(ctx); err != nil {
					return err
				}

				return nil
			})

			framework.Serve()
		},
	}

	rootCmd.AddCommand(
		cmds.InstallCommand(&cmds.InstallCommandConfig{
			PluginName: "portmaster-plugin-dnscrypt",
			Types: []shared.PluginType{
				shared.PluginTypeResolver,
			},
		}),
	)

	if err := rootCmd.Execute(); err != nil {
		hclog.L().Error(err.Error())
		os.Exit(1)
	}
}
