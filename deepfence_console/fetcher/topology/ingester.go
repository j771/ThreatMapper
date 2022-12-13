package topology

import (
	"fmt"

	"github.com/deepfence/fetcher_api_server/types"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
)

type TopologyClient struct {
	driver neo4j.Driver
}

func NewTopologyClient() *TopologyClient {
	driver, err := neo4j.NewDriver("bolt://neo4j-db:7687", neo4j.BasicAuth("neo4j", "password", ""))

	if err != nil {
		return nil
	}

	nc := &TopologyClient{
		driver: driver,
	}

	return nc
}
func (tc *TopologyClient) Close() {
	tc.driver.Close()
}

func (tc *TopologyClient) AddCompliances(cs []types.ComplianceDoc) error {
	session, err := tc.driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Close()

	if _, err = tx.Run("UNWIND $batch as row MERGE (n:Compliance{node_id:row.node_id, test_number:row.test_number}) SET n+= row", map[string]interface{}{"batch": types.CompliancesToMaps(cs)}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Compliance) MERGE (m:ComplianceScan{node_id: n.scan_id, time_stamp: timestamp()}) MERGE (m) -[:DETECTED]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Compliance) MERGE (m:ComplianceScan{node_id: n.scan_id}) MERGE (l:KCluster{node_id: n.kubernetes_cluster_id}) MERGE (m) -[:SCANNED]-> (l)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Node) WHERE n.kubernetes_cluster_id IS NOT NULL AND n.kubernetes_cluster_id <> '' MERGE (m:KCluster{node_id:n.kubernetes_cluster_id}) MERGE (m) -[:KHOSTS]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	return tx.Commit()
}

func (tc *TopologyClient) AddCVEs(cs []types.DfCveStruct) error {
	session, err := tc.driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Close()

	if _, err = tx.Run("UNWIND $batch as row MERGE (n:Cve{node_id:row.cve_id}) SET n+= row", map[string]interface{}{"batch": types.CVEsToMaps(cs)}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Cve) MERGE (m:CveScan{node_id: n.scan_id, host_name:n.host_name, time_stamp: timestamp()}) MERGE (m) -[:DETECTED]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CveScan) MERGE (m:Node{node_id: n.host_name}) MERGE (n) -[:SCANNED]-> (m)", map[string]interface{}{}); err != nil {
		return err
	}

	return tx.Commit()
}

func (tc *TopologyClient) AddSecrets(cs []map[string]interface{}) error {
	session, err := tc.driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Close()

	secrets := []map[string]interface{}{}
	for _, i := range cs {
		secret := map[string]interface{}{}
		match := i["Match"].(map[string]interface{})
		severity := i["Severity"].(map[string]interface{})
		rule := i["Rule"].(map[string]interface{})

		for k, v := range i {
			if k == "Match" || k == "Severity" || k == "Rule" {
				continue
			}
			secret[k] = v
		}

		for k, v := range rule {
			secret[k] = v
		}
		for k, v := range severity {
			secret[k] = v
		}
		for k, v := range match {
			secret[k] = v
		}
		secret["rule_id"] = fmt.Sprintf("%v:%v", rule["id"], i["host_name"])
		secrets = append(secrets, secret)
	}

	if _, err = tx.Run("UNWIND $batch as row MERGE (n:Secret{node_id:row.rule_id}) MERGE (m:SecretScan{node_id: row.scan_id, host_name: row.host_name, time_stamp: timestamp()}) MERGE (m) -[:DETECTED]-> (n) SET n+= row", map[string]interface{}{"batch": secrets}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:SecretScan) MERGE (m:Node{node_id: n.host_name}) MERGE (n) -[:SCANNED]-> (m)", map[string]interface{}{}); err != nil {
		return err
	}

	return tx.Commit()
}

func (tc *TopologyClient) AddCloudResources(cs []types.CloudResource) error {
	fmt.Println("test everything here", cs)
	session, err := tc.driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Close()

	if _, err = tx.Run("UNWIND $batch as row MERGE (m:CloudResource{node_id:row.arn, resource_type:row.resource_id}) SET m+=row", map[string]interface{}{"batch": types.ResourceToMaps(cs)}); err != nil {
		return err
	}

	fmt.Println("test everything here", cs)
	if _, err = tx.Run("UNWIND $batch as row MERGE (m:CloudResource{node_id:row.arn, resource_type:row.resource_id}) SET m+=row WITH row UNWIND apoc.convert.fromJsonList(row.security_groups) as group WITH group, row WHERE group IS NOT NULL AND  row.resource_id = 'aws_ec2_instance' AND group.GroupId IS NOT NULL MERGE (n:SecurityGroup{node_id:group.GroupId, name:group.GroupName}) MERGE (m:CloudResource{node_id:row.arn, resource_type:row.resource_id}) MERGE (n)-[:SECURED]->(m)", map[string]interface{}{"batch": types.ResourceToMaps(cs)}); err != nil {
		return err
	}

	return tx.Commit()
}

func (tc *TopologyClient) AddCloudCompliances(cs []types.CloudComplianceDoc) error {
	session, err := tc.driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Close()

	if _, err = tx.Run("UNWIND $batch as row MERGE (n:CloudCompliance{resource:row.resource, reason: row.reason}) MERGE (m:CloudResource{node_id:row.resource}) MERGE (n) -[:SCANNED]-> (m) SET n+= row", map[string]interface{}{"batch": types.CloudCompliancesToMaps(cs)}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudCompliance) MERGE (m:CloudComplianceScan{node_id: n.scan_id, time_stamp: timestamp()}) MERGE (m) -[:DETECTED]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	return tx.Commit()
}

func (tc *TopologyClient) LinkNodesWithCloudResources() error {
	session, err := tc.driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Close()

	if _, err = tx.Run("match (n) -[r:IS]-> (m) delete r", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("match (n:Node) WITH apoc.convert.fromJsonMap(n.cloud_metadata) as map, n WHERE map.label = 'AWS' WITH map.id as id, n match (m:CloudResource) where m.resource_type = 'aws_ec2_instance' and m.instance_id = id MERGE (n) -[:IS]-> (m)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("match (n:Node) WITH apoc.convert.fromJsonMap(n.cloud_metadata) as map, n WHERE map.label = 'GCP' WITH map.hostname as hostname, n match (m:CloudResource) where m.resource_type = 'gcp_compute_instance' and m.hostname = hostname MERGE (n) -[:IS]-> (m)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("match (n:Node) WITH apoc.convert.fromJsonMap(n.cloud_metadata) as map, n WHERE map.label = 'AZURE' WITH map.vmId as vm, n match (m:CloudResource) where m.resource_type = 'azure_compute_virtual_machine' and m.arn = vm MERGE (n) -[:IS]-> (m)", map[string]interface{}{}); err != nil {
		return err
	}

	return tx.Commit()
}

func (tc *TopologyClient) ComputeThreatGraph() error {
	session, err := tc.driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Close()

	if _, err = tx.Run("MATCH (s:CveScan) -[:SCANNED]-> (m) WITH max(s.time_stamp) as most_recent, m MATCH (s:CveScan {time_stamp: most_recent})-[:DETECTED]->(c:Cve) WITH m, count(distinct c) as num_cve SET m.num_cve = num_cve", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (s:SecretScan) -[:SCANNED]-> (m) WITH max(s.time_stamp) as most_recent, m MATCH (s:SecretScan {time_stamp: most_recent})-[:DETECTED]->(c:Secret) WITH m, count(distinct c) as num_secrets SET m.num_secrets = num_secrets", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (s:ComplianceScan) -[:SCANNED]-> (m) WITH max(s.time_stamp) as most_recent, m MATCH (s:ComplianceScan {time_stamp: most_recent})-[:DETECTED]->(c:Compliance) WITH m, count(distinct c) as num_compliance SET m.num_compliance = num_compliance", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Node) SET n.sum_cve = COALESCE(n.num_cve, 0), n.sum_secrets = COALESCE(n.num_secrets, 0), n.sum_compliance = COALESCE(n.num_compliance, 0);", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Node) -[:CONNECTED]->(m:Node) SET n.sum_cve = COALESCE(n.sum_cve, 0) + COALESCE(m.sum_cve, m.num_cve, 0), n.sum_secrets = COALESCE(n.sum_secrets, 0) + COALESCE(m.sum_secrets, m.num_secrets, 0), n.sum_compliance = COALESCE(n.sum_compliance, 0) + COALESCE(m.sum_compliance, m.num_compliance, 0);", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Node {node_id:'in-the-internet'})-[d:CONNECTS*]->(m:Node) with SIZE(d) as depth, m with min(depth) as min_depth, m SET m.depth = min_depth", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Node) SET n.num_cve = COALESCE(n.num_cve, 0), n.num_secrets = COALESCE(n.num_secrets, 0), n.num_compliance = COALESCE(n.num_compliance, 0);", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Node) SET n.sum_cve = n.num_cve, n.sum_secrets = n.num_secrets, n.sum_compliance = n.num_compliance;", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:Node) -[:CONNECTS]->(m:Node) SET n.sum_cve = COALESCE(n.sum_cve, 0) + COALESCE(m.sum_cve, m.num_cve, 0), n.sum_secrets = COALESCE(n.sum_secrets, 0) + COALESCE(m.sum_secrets, m.num_secrets, 0), n.sum_compliance = COALESCE(n.sum_compliance, 0) + COALESCE(m.sum_compliance, m.num_compliance, 0);", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_vpc_security_group_rule'})  MATCH (m:SecurityGroup{node_id: n.group_id}) -[:SECURED]-> (z:CloudResource{resource_type:'aws_ec2_instance'}) WHERE n.is_egress <> true   MERGE (k:Node {node_id:'in-the-internet'})  MERGE (k)-[:PUBLIC]->(z)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_vpc_security_group_rule'}) MATCH (m:SecurityGroup{node_id: n.group_id})-[:SECURED]-> (z:CloudResource{resource_type:'aws_ec2_instance'}) WHERE n.is_egress = true   MERGE (k:Node {node_id:'out-the-internet'})  MERGE (z)-[:PUBLIC]->(k)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (m:CloudResource{resource_type:'aws_ecs_service'})  MATCH (k:CloudResource{resource_type:'aws_ecs_task'}) MATCH (n:CloudResource{resource_type:'aws_ecs_task_definition',node_id:substring(m.task_definition,1,size(m.task_definition)-2)}) WITH apoc.convert.fromJsonMap(m.network_configuration) as map,m,n,k WHERE n.node_id=k.task_definition_arn AND map is not null AND  map.AwsvpcConfiguration.AssignPublicIp = 'ENABLED' MERGE (p:Node {node_id:'in-the-internet'})  MERGE (p) -[:PUBLIC]-> (k)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (k:CloudResource{resource_type:'aws_ecs_task'})  WITH apoc.convert.fromJsonList(k.containers) as containers,k  UNWIND containers as container  MERGE  (n:Container{node_id:container.container_runtime_id,docker_container_id:container.container_runtime_id,docker_container_name:container.container_name})   MERGE (k) -[:USES]-> (n) MERGE  (l:ContainerImage{node_id:container.image_uri}) MERGE (n) -[:USES]-> (l)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_s3_bucket', bucket_policy_is_public: true })  MATCH (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run(
		"MATCH (t)-[r:PUBLIC]->(k:CloudResource{resource_type:'aws_ec2_instance' } ) "+
			"MATCH (n:CloudResource{resource_type:'aws_iam_role'}) WITH apoc.convert.fromJsonList(n.instance_profile_arns) "+
			"as instance_arns,k,n WHERE k.iam_instance_profile_arn IN instance_arns "+
			"WITH apoc.convert.fromJsonList(n.attached_policy_arns) as attached_policy_arns,k,n   UNWIND attached_policy_arns as policy_arn  "+
			"MATCH (z:CloudResource{resource_type:'aws_iam_policy' }) where  z.node_id = policy_arn  WITH apoc.convert.fromJsonMap(z.policy_std) "+
			"as policy,z,k,n  UNWIND policy.Statement as pol_statement WITH apoc.convert.fromJsonList(pol_statement.Action) as pol_actions,z,k,n  "+
			"UNWIND pol_actions as pol_act where pol_act =~ '.*S3.*'  and pol_statement.Effect <> 'Deny'   "+
			" MATCH (c:CloudResource{resource_type:'aws_s3_bucket' } )   with WITH apoc.convert.fromJsonMap(c.policy) as buc_pol,z,n,k,c"+
			"UNWIND buc_pol.Statement as buc_pol_statement"+
			"WITH apoc.convert.fromJsonList(buc_pol_statement.Action) as buc_pol_actions,z,k,n,C  UNWIND buc_pol_actions as buc_pol_act"+
			" where buc_pol_act <>~ '.*InstanceProfile.*'  or pol_statement.Effect <> 'Deny'"+
			" MERGE (k) -[:COMMUNICATES]-> (c)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_lambda_function' }) WITH apoc.convert.fromJsonMap(n.policy_std) as policy,n UNWIND policy.Statement as pol_statement   MATCH (p:Node {node_id:'in-the-internet'})   where pol_statement.Principal.AWS = '*' and pol_statement.Effect = 'Allow' MERGE (p) -[:PUBLIC]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_s3_bucket' }) WITH apoc.convert.fromJsonMap(n.event_notification_configuration) as eventConfig,n  UNWIND eventConfig.LambdaFunctionConfigurations AS envconf MATCH (p:CloudResource{resource_type:'aws_lambda_function' , arn: envconf.LambdaFunctionArn }) MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}
	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_ec2_classic_load_balancer', scheme : 'internet_facing' }) MATCH (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_ec2_network_load_balancer', scheme : 'internet_facing' }) MATCH (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_ec2_application_load_balancer', scheme : 'internet_facing' }) MATCH (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_opensearch_domain' }) where n.vpc_options IS NULL MATCH (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_rds_db_cluster'}) WITH apoc.convert.fromJsonMap(n.vpc_security_groups) as vpc_security_groups,n where vpc_security_groups.VpcSecurityGroupId.is_egress IS NOT NULL and  vpc_security_groups.VpcSecurityGroupId.cidr_ipv4= '0.0.0.0/0'  MERGE (p:Node {node_id:'out-the-internet'})   MERGE (n) -[:PUBLIC]-> (p) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'aws_rds_db_cluster'}) WITH apoc.convert.fromJsonMap(n.vpc_security_groups) as vpc_security_groups,n where vpc_security_groups.VpcSecurityGroupId.is_egress IS  NULL and  vpc_security_groups.VpcSecurityGroupId.cidr_ipv4= '0.0.0.0/0'  MATCH (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'azure_storage_account', allow_blob_public_access: true }) MERGE (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'azure_storage_account', allow_blob_public_access: true}) MATCH (p:CloudResource{resource_type:'azure_storage_blob', storage_account_name: n.node_id })    MERGE (p) -[:PUBLIC]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'azure_storage_account', allow_blob_public_access: true}) MATCH (p:CloudResource{resource_type:'azure_storage_table', storage_account_name: n.node_id })    MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'azure_storage_account', allow_blob_public_access: true}) MATCH (p:CloudResource{resource_type:'azure_log_profile', storage_account_name: n.node_id })    MERGE (p) -[:PUBLIC]-> (n)", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'azure_mysql_server', public_network_access: 'Enabled'})   MERGE (p:Node {node_id:'in-the-internet'})   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'azure_storage_container'}) MATCH (p:CloudResource{resource_type:'azure_log_profile', storage_account_name: n.node_id }) WHERE (n.public_access IS NOT NULL) OR (n.public_access <> '')   MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'azure_compute_virtual_machine' })  WITH apoc.convert.fromJsonList(n.public_ips) as public_ips,n  WHERE (public_ips IS NOT NULL) OR (size(public_ips)>0) MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'gcp_compute_instance' })  WITH apoc.convert.fromJsonList(n.network_interfaces) as network_interfaces,n  UNWIND network_interfaces AS network_interface  MERGE (p:Node {node_id:'in-the-internet'}) WHERE (network_interface IS NOT NULL) AND (network_interface.accessConfigs IS NOT NULL) AND (network_interface.accessConfigs.accessConfigs.natIP IS NOT NULL) MERGE (p) -[:PUBLIC]-> (n) where ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'gcp_storage_bucket' })  WITH apoc.convert.fromJsonMap(n.iam_policy) as policy,n  UNWIND iam_policy.bindings AS binding   WHERE (binding =~ '.*allAuthenticatedUsers.*' ) AND (binding =~ '.*allUsers.*')  MERGE (p:Node {node_id:'in-the-internet'}) MERGE (p) -[:PUBLIC]-> (n) where ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'gcp_sql_database_instance' })  WITH apoc.convert.fromJsonMap(n.ip_configuration) as ip_config,n UNWIND ip_config.authorizedNetworks as network  WHERE (network.value = '0.0.0.0/0') MERGE (p:Node {node_id:'in-the-internet'})  MERGE (p) -[:PUBLIC]-> (n) ", map[string]interface{}{}); err != nil {
		return err
	}

	if _, err = tx.Run("MATCH (n:CloudResource{resource_type:'gcp_cloudfunctions_function', ingress_settings: 'ALLOW_ALL' })    MERGE (p:Node {node_id:'in-the-internet'}) MERGE (p) -[:PUBLIC]-> (n)) where ", map[string]interface{}{}); err != nil {
		return err
	}

	return tx.Commit()
}

const (
	CLOUD_AWS     = "aws"
	CLOUD_AZURE   = "azure"
	CLOUD_GCP     = "gcp"
	CLOUD_PRIVATE = "others"
)

var CLOUD_ALL = [...]string{CLOUD_AWS, CLOUD_AZURE, CLOUD_GCP, CLOUD_PRIVATE}

func (tc *TopologyClient) GetThreatGraph() (ThreatGraph, error) {
	aggreg, err := tc.GetRawThreatGraph()
	if err != nil {
		return ThreatGraph{}, err
	}

	all := ThreatGraph{}
	for _, cp := range CLOUD_ALL {
		resources := []ThreatNodeInfo{}
		node_info := aggreg[cp].getNodeInfos()
		depths := aggreg[cp].nodes_depth
		if _, has := depths[1]; !has {
			goto end
		}
		for _, root := range depths[1] {
			visited := map[int64]struct{}{}
			attack_paths := build_attack_paths(aggreg[cp], root, visited)
			paths := [][]string{}
			for _, Attack_path := range attack_paths {
				path := []string{}
				for i := range Attack_path {
					index := int64(len(Attack_path)-1) - int64(i)
					path = append(path, node_info[index].Id)
				}
				paths = append(paths, append([]string{"The Internet"}, path...))
				entry := ThreatNodeInfo{
					Label:               node_info[int64(len(Attack_path)-1)].Label,
					Id:                  node_info[int64(len(Attack_path)-1)].Id,
					Nodes:               node_info[int64(len(Attack_path)-1)].Nodes,
					Vulnerability_count: node_info[int64(len(Attack_path)-1)].Vulnerability_count,
					Secrets_count:       node_info[int64(len(Attack_path)-1)].Secrets_count,
					Compliance_count:    node_info[int64(len(Attack_path)-1)].Compliance_count,
					Count:               node_info[int64(len(Attack_path)-1)].Count,
					Node_type:           node_info[int64(len(Attack_path)-1)].Node_type,
					Attack_path:         paths,
				}
				resources = append(resources, entry)
			}
		}
	end:
		all[cp] = ProviderThreatGraph{
			Resources:           resources,
			Compliance_count:    0,
			Secrets_count:       0,
			Vulnerability_count: 0,
		}
	}

	return all, nil
}

func build_attack_paths(paths AttackPaths, root int64, visited map[int64]struct{}) [][]int64 {
	if _, has := visited[root]; has {
		return [][]int64{}
	}
	visited[root] = struct{}{}
	if _, has := paths.nodes_data[root]; !has {
		return [][]int64{}
	}
	if _, has := paths.nodes_tree[root]; !has {
		return [][]int64{{root}}
	}
	res := [][]int64{{}}
	for _, edge := range paths.nodes_tree[root] {
		edge_paths := build_attack_paths(paths, edge, visited)
		for _, edge_path := range edge_paths {
			res = append(res, append([]int64{root}, edge_path...))
		}
	}
	if len(res) == 0 {
		return [][]int64{{root}}
	}
	return res
}

func (tc *TopologyClient) GetRawThreatGraph() (map[string]AttackPaths, error) {
	session, err := tc.driver.Session(neo4j.AccessModeRead)

	if err != nil {
		return nil, err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return nil, err
	}
	defer tx.Close()

	all := map[string]AttackPaths{}
	for _, cloud_provider := range CLOUD_ALL {
		var res neo4j.Result
		if cloud_provider != CLOUD_PRIVATE {
			if res, err = tx.Run(`
				CALL apoc.nodes.group(['Node'], ['node_type', 'depth',
				'cloud_provider'], [{`+"`*`"+`: 'count', sum_cve: 'sum', sum_secrets: 'sum', sum_compliance: 'sum',
				node_id:'collect', num_cve: 'collect', num_secrets:'collect', num_compliance:'collect'},{`+"`*`"+`: 'count'}], {selfRels: false})
				YIELD node, relationships
				WHERE apoc.any.property(node, 'depth') IS NOT NULL
				AND apoc.any.property(node, 'cloud_provider') = '`+cloud_provider+`'
				RETURN node, relationships
				`, map[string]interface{}{}); err != nil {
			}
		} else {
			if res, err = tx.Run(`
				CALL apoc.nodes.group(['Node'], ['node_type', 'depth',
				'cloud_provider'], [{`+"`*`"+`: 'count', sum_cve: 'sum', sum_secrets: 'sum', sum_compliance: 'sum',
				node_id:'collect', num_cve: 'collect', num_secrets:'collect', num_compliance:'collect'},{`+"`*`"+`: 'count'}], {selfRels: false})
				YIELD node, relationships
				WHERE apoc.any.property(node, 'depth') IS NOT NULL
				AND NOT apoc.any.property(node, 'cloud_provider') IN ['aws', 'gcp', 'azure']
				AND apoc.any.property(node, 'cloud_provider') <> 'internet'
				RETURN node, relationships
				`, map[string]interface{}{}); err != nil {
			}
		}

		if err != nil {
			return nil, err
		}

		records, err := res.Collect()
		if err != nil {
			return nil, err
		}

		nodes_tree := map[int64][]int64{}
		nodes_data := map[int64]AttackPathData{}
		nodes_depth := map[int64][]int64{}
		for _, record := range records {
			record_node, _ := record.Get("node")
			record_relationships, _ := record.Get("relationships")
			node := record_node.(dbtype.Node)
			node_datum := record2struct(node)
			nodes_data[node.Id] = node_datum

			for _, rel_node := range record_relationships.([]interface{}) {
				rel := rel_node.(dbtype.Relationship)
				nodes_tree[node.Id] = append(nodes_tree[node.Id], rel.EndId)

			}
			nodes_depth[node_datum.depth] = append(nodes_depth[node_datum.depth], node.Id)
		}

		all[cloud_provider] = AttackPaths{
			nodes_tree:  nodes_tree,
			nodes_data:  nodes_data,
			nodes_depth: nodes_depth,
		}
	}

	return all, nil
}

type AttackPaths struct {
	nodes_tree  map[int64][]int64
	nodes_data  map[int64]AttackPathData
	nodes_depth map[int64][]int64
}

func record2struct(node dbtype.Node) AttackPathData {

	record := node.Props
	Node_type, _ := record["node_type"]
	depth, _ := record["depth"]
	cloud_provider, _ := record["cloud_provider"]
	sum_sum_cve, _ := record["sum_sum_cve"]
	sum_sum_secrets, _ := record["sum_sum_secrets"]
	sum_sum_compliance, _ := record["sum_sum_compliance"]
	node_count, _ := record["count_*"]
	collect_node_id_, _ := record["collect_node_id"]
	collect_num_cve_, _ := record["collect_num_cve"]
	collect_num_secrets_, _ := record["collect_num_secrets"]
	collect_num_compliance_, _ := record["collect_num_compliance"]

	collect_node_id := []string{}
	for _, v := range collect_node_id_.([]interface{}) {
		collect_node_id = append(collect_node_id, v.(string))
	}

	collect_num_cve := []int64{}
	for _, v := range collect_num_cve_.([]interface{}) {
		collect_num_cve = append(collect_num_cve, v.(int64))
	}

	collect_num_secrets := []int64{}
	for _, v := range collect_num_secrets_.([]interface{}) {
		collect_num_secrets = append(collect_num_secrets, v.(int64))
	}

	collect_num_compliance := []int64{}
	for _, v := range collect_num_compliance_.([]interface{}) {
		collect_num_compliance = append(collect_num_compliance, v.(int64))
	}

	return AttackPathData{
		Node_type:              Node_type.(string),
		cloud_provider:         cloud_provider.(string),
		depth:                  depth.(int64),
		sum_sum_cve:            sum_sum_cve.(int64),
		sum_sum_secrets:        sum_sum_secrets.(int64),
		sum_sum_compliance:     sum_sum_compliance.(int64),
		node_count:             node_count.(int64),
		collect_node_id:        collect_node_id,
		collect_num_cve:        collect_num_cve,
		collect_num_secrets:    collect_num_secrets,
		collect_num_compliance: collect_num_compliance,
	}
}

type AttackPathData struct {
	identity               int64
	Node_type              string
	cloud_provider         string
	depth                  int64
	sum_sum_cve            int64
	sum_sum_secrets        int64
	sum_sum_compliance     int64
	node_count             int64
	collect_node_id        []string
	collect_num_cve        []int64
	collect_num_secrets    []int64
	collect_num_compliance []int64
}

var container_id = 0
var host_id = 0

func (ap AttackPaths) getNodeInfos() map[int64]ThreatNodeInfo {
	res := map[int64]ThreatNodeInfo{}
	for _, v := range ap.nodes_data {
		var Label, Id string
		switch v.Node_type {
		case "host":
			Label = "Compute Instance"
			Id = fmt.Sprintf("%v-host-%v", v.cloud_provider, host_id)
			host_id += 1
		case "container":
			Label = "Container"
			Id = fmt.Sprintf("%v-container-%v", v.cloud_provider, container_id)
			container_id += 1
		default:
			Label = "The Internet"
			Id = "The Internet"
		}
		Nodes := map[string]NodeInfo{}
		for i, Node_id := range v.collect_node_id {
			Nodes[Node_id] = NodeInfo{
				Node_id:               v.collect_node_id[i],
				Image_name:            "",
				Name:                  Node_id,
				Vulnerability_count:   v.collect_num_cve[i],
				Vulnerability_scan_id: "",
				Secrets_count:         v.collect_num_secrets[i],
				Secrets_scan_id:       "",
				Compliance_count:      v.collect_num_compliance[i],
				Compliance_scan_id:    "",
			}
		}
		res[v.identity] = ThreatNodeInfo{
			Label:               Label,
			Id:                  Id,
			Nodes:               Nodes,
			Vulnerability_count: v.sum_sum_cve,
			Secrets_count:       v.sum_sum_secrets,
			Compliance_count:    v.sum_sum_compliance,
			Count:               int64(len(v.collect_node_id)),
			Node_type:           v.Node_type,
			Attack_path:         [][]string{},
		}
	}
	return res
}

type ThreatGraph map[string]ProviderThreatGraph

type ProviderThreatGraph struct {
	Resources           []ThreatNodeInfo `json:"resources"`
	Compliance_count    int64            `json:"compliance_count"`
	Secrets_count       int64            `json:"secrets_count"`
	Vulnerability_count int64            `json:"vulnerability_count"`
}

type ThreatNodeInfo struct {
	Label string              `json:"label"`
	Id    string              `json:"id"`
	Nodes map[string]NodeInfo `json:"nodes"`

	Vulnerability_count int64 `json:"vulnerability_count"`
	Secrets_count       int64 `json:"secrets_count"`
	Compliance_count    int64 `json:"compliance_count"`
	Count               int64 `json:"count"`

	Node_type string `json:"node_type"`

	Attack_path [][]string `json:"attack_path"`
}

type NodeInfo struct {
	Node_id               string `json:"node_id"`
	Image_name            string `json:"image_name"`
	Name                  string `json:"name"`
	Vulnerability_count   int64  `json:"vulnerability_count"`
	Vulnerability_scan_id string `json:"vulnerability_scan_id"`
	Secrets_count         int64  `json:"secrets_count"`
	Secrets_scan_id       string `json:"secrets_scan_id"`
	Compliance_count      int64  `json:"compliance_count"`
	Compliance_scan_id    string `json:"compliance_scan_id"`
}
