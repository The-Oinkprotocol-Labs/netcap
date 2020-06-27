package transform

import (
	maltego "github.com/dreadl0ck/netcap/maltego"
	"github.com/dreadl0ck/netcap/types"
	"strings"
)

func ToProducts() {
	maltego.SoftwareTransform(
		nil,
		func(lt maltego.LocalTransform, trx *maltego.MaltegoTransform, soft *types.Software, min, max uint64, profilesFile string, mac string, ipaddr string) {

			val := soft.Vendor + " " + soft.Product + " " + soft.Version
			if len(soft.SourceName) > 0 {
				if soft.SourceName == "Generic version harvester" {
					if len(val) == 0 {
						val = maltego.EscapeText(soft.SourceData)
					} else {
						val += "\n" + maltego.EscapeText(soft.SourceData)
					}
				}
				val += "\n" + soft.SourceName
			}
			for i, f := range soft.Flows {
				if i == 3 {
					val += "\n..."
					break
				}
				val += "\n" + f
			}

			val = maltego.EscapeText(val)

			ent := trx.AddEntity("netcap.Software", val)
			ent.SetType("netcap.Software")
			ent.SetValue(val)

			ent.AddProperty("timestamp", "Timestamp", "strict", soft.Timestamp)
			ent.AddProperty("vendor", "Vendor", "strict", maltego.EscapeText(soft.Vendor))
			ent.AddProperty("product", "Product", "strict", maltego.EscapeText(soft.Product))
			ent.AddProperty("version", "Version", "strict", maltego.EscapeText(soft.Version))
			ent.AddProperty("flows", "Flows", "strict", strings.Join(soft.Flows, " | "))
			ent.AddProperty("sourcename", "SourceName", "strict", soft.SourceName)
			ent.AddProperty("sourcedata", "SourceData", "strict", maltego.EscapeText(soft.SourceData))
			ent.AddProperty("notes", "Notes", "strict", maltego.EscapeText(soft.Notes))

			ent.SetLinkColor("#000000")
			//ent.SetLinkThickness(maltego.GetThickness(uint64(count), min, max))
		},
	)
}
