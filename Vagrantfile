
# TODO: FreeBSD and OmniOS box

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
	config.vm.box = "chef/ubuntu-14.04"
	config.ssh.forward_agent = true

    config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/bakins/zfsd", create: true
    config.vm.network "forwarded_port", guest: 9373, host: 9373, auto_correct: true

    config.vm.provision "shell", privileged: true, inline: <<EOF

chown -R vagrant /home/vagrant/go

if [ ! -x /sbin/zpool ]; then
    apt-get update
    apt-get -y install software-properties-common linux-headers-\$(uname -r)
    add-apt-repository ppa:zfs-native/stable
    apt-get update
    apt-get -y install ubuntu-zfs
    modprobe zfs
    echo zfs >> /etc/modules
fi

if [ ! -x /usr/local/go/bin/go ]; then
     apt-get update
     apt-get -y install curl git-core mercurial
     cd /tmp
     curl -s -L -O http://golang.org/dl/go1.3.3.linux-amd64.tar.gz
     tar -C /usr/local -zxf go1.3.3.linux-amd64.tar.gz
     rm /tmp/go1.3.3.linux-amd64.tar.gz
fi

if [ ! -x /usr/local/bin/jq ]; then
    cd /tmp
    curl -s -L -O http://stedolan.github.io/jq/download/linux64/jq
    mv jq /usr/local/bin
    chmod 0555 /usr/local/bin/jq
fi

zpool list testing
if [ $? -ne 0 ]; then
    for i in {0..4}; do
        truncate -s 2G /root/testing-$i.img
     done
    zpool create testing raidz1 /root/testing-*.img
    zfs set compression=lz4 testing
    zfs create testing/A
    zfs create testing/B
    zfs create -V 8192 testing/C
    zfs snapshot testing/A@123456789
    zfs clone testing/A@123456789 testing/foo
fi

cat <<EOS > /etc/profile.d/go.sh
GOPATH=\\$HOME/go
export GOPATH
PATH=\\$GOPATH/bin:\\$PATH:/usr/local/go/bin
export PATH
EOS

cat << END > /etc/sudoers.d/go
Defaults env_keep += "GOPATH"
END

mkdir -p /home/vagrant/.ssh
cat << END  > /home/vagrant/.ssh/config
Host github.com
    StrictHostKeyChecking no
END

chown -R vagrant /home/vagrant/.ssh

EOF

end
