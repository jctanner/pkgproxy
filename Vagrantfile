Vagrant.configure("2") do |config|
  #config.vm.synced_folder ".", "/vagrant", type: "nfs", nfs_udp: false
  config.vm.synced_folder ".", "/vagrant", type: "rsync"

  # https://github.com/vagrant-libvirt/vagrant-libvirt/issues/760
  # vagrant-libvirt doesn't support macvtp+wifi so we just forward
  # these ports to get around that
  config.vm.network "forwarded_port", guest: 80, host: 80
  config.vm.network "forwarded_port", guest: 43, host: 443

  config.vm.define "pkgproxy" do |dev|
    dev.vm.box = "generic/debian11"
    dev.vm.hostname = "pkgproxy.lab.net"

    dev.vm.provider :libvirt do |libvirt|
      libvirt.cpus = 2
      libvirt.memory = 1000
      libvirt.machine_virtual_size = 20
    end
  end

  config.vm.provision "shell", inline: <<-SHELL
    apt -y update
    apt -y install golang
  SHELL

end
