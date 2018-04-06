require 'rspec'
require 'bosh/template/test'
require 'yaml'
require 'json'

module Bosh::Template::Test
  describe 'cni-wrapper-plugin.conf.erb' do
    describe 'template rendering' do
      let(:release_path) {File.join(File.dirname(__FILE__), '../..')}
      let(:release) {ReleaseDir.new(release_path)}

      describe 'cni job' do
        let(:job) {release.job('cni')}
        describe 'config/cni/cni-wrapper-plugin.conf' do
          let(:merged_manifest_properties) { {
            'iptables_logging' => true,
            'iptables_accepted_udp_logs_per_sec' => 15,
            'iptables_denied_logs_per_sec' => 14,
            'overlay_network' => '10.255.0.0/16',
            'proxy_port' => 6868,
            'dns_servers' => ['8.8.8.8'],
            'silk_daemon' => {
              'listen_port' => 3636,
            },
            'mtu' => 1250,
            'rate' => 42,
            'burst' => 43
          } }

          let(:links) { [
            Link.new(
              name: 'cf_network',
              instances: [LinkInstance.new()],
              properties: {
                'network' => '10.255.0.0/16',
                'subnet_prefix_length' => 24
              }
            )
          ]}

          let(:template) {job.template('config/cni/cni-wrapper-plugin.conf')}
          let(:instance_spec) {
            InstanceSpec.new(
              networks: networks,
              ip: '192.168.4.2'
          )}

          let(:networks) do
            {
              'network1' => {'ip' => '1.2.3.4'},
              'network2' => {'ip' => '2.3.4.5'}
            }
          end

          subject(:parsed_conf) do
            JSON.parse(template.render(merged_manifest_properties, spec: instance_spec, consumes: links))
          end

          it 'renders the config' do
            expect(parsed_conf).to eq({
              'name' => 'cni-wrapper',
              'type' => 'cni-wrapper-plugin',
              'cniVersion' => '0.3.1',
              'datastore' => '/var/vcap/data/container-metadata/store.json',
              'iptables_lock_file' => '/var/vcap/data/garden-cni/iptables.lock',
              'instance_address' => '192.168.4.2',
              'underlay_ips' => ['1.2.3.4', '2.3.4.5'],
              'overlay_network' => '10.255.0.0/16',
              'proxy_port' => 6868,
              'iptables_asg_logging' => true,
              'iptables_c2c_logging' => true,
              'iptables_denied_logs_per_sec' => 14,
              'iptables_accepted_udp_logs_per_sec' => 15,
              'ingress_tag' => 'ffff0000',
              'vtep_name' => 'silk-vtep',
              'dns_servers' => ['8.8.8.8'],
              'delegate' => {
                'cniVersion' => '0.3.1',
                'name' => 'silk',
                'type' => 'silk-cni',
                'daemonPort' => 3636,
                'dataDir' => '/var/vcap/data/host-local',
                'datastore' => '/var/vcap/data/silk/store.json',
                'mtu' => 1200,
                'bandwidthLimits'=> {
                  'rate' => 42 * 1024,
                  'burst' => 43 * 1024
                }
               }
            })
          end

          context 'when multiple networks are present' do
            context 'and the temporary_underlay_interface_ips property is not set' do
              it 'renders the bosh network ips into the manifest' do
                expect(parsed_conf['underlay_ips']).to contain_exactly('1.2.3.4', '2.3.4.5')
              end
            end

            context 'and the temporary_underlay_interface_ips property is set' do
              let(:merged_manifest_properties) { {'temporary' => {'underlay_interface_ips' => ['5.5.5.5', '8.8.8.8'] }} }

              it 'renders the given ips into the manifest' do
                expect(parsed_conf['underlay_ips']).to contain_exactly('5.5.5.5', '8.8.8.8')
              end
            end
          end
        end
      end
    end
  end
end
