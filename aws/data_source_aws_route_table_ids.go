package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsRouteTableIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRouteTableIDsRead,

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": ec2CustomFiltersSchema(),
			"tags":   tagsSchemaComputed(),
			"ids": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsRouteTableIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	req := &ec2.DescribeRouteTablesInput{}
	vpcId, vpcIdOk := d.GetOk("vpc_id")
	subnetId, subnetIdOk := d.GetOk("subnet_id")
	rtbId, rtbIdOk := d.GetOk("route_table_id")
	tags, tagsOk := d.GetOk("tags")
	filter, filterOk := d.GetOk("filter")

	attrMap := map[string]string{}
	if vpcIdOk {
		attrMap["vpc-id"] = vpcId.(string)
	}
	if subnetIdOk {
		attrMap["association.subnet-id"] = subnetId.(string)
	}
	if rtbIdOk {
		attrMap["route-table-id"] = rtbId.(string)
	}

	req.Filters = buildEC2AttributeFilterList(attrMap)
	if tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			tagsFromMap(tags.(map[string]interface{})),
		)...)
	}
	if filterOk {
		req.Filters = append(req.Filters, buildEC2CustomFilterList(
			filter.(*schema.Set),
		)...)
	}

	log.Printf("[DEBUG] Describe Route Tables %v\n", req)
	resp, err := conn.DescribeRouteTables(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.RouteTables) == 0 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	rts := make([]string, 0)
	for _, rt := range resp.RouteTables {
		rts = append(rts, *rt.RouteTableId)
	}

	d.SetId(resource.UniqueId())
	d.Set("ids", rts)

	return nil
}
