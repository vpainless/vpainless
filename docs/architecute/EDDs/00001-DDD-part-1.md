# EDD 00001: Domain Driven Design - Part 1

## Context

This document is an exploration of requirement and features. It's objective is to shape an initial draft of potential bounded contexts, subdomains and ubiquitous language.

## Ubiquitous Language

- **User**: Anyone that registers in our application, using a username and password.
- **Group**: Any user that's not part of any group can create a group, they'll become the admin of the group.
- **Admin**: The use that has created the group, of has been promoted by another group admin.
- **Client**: Any user made by an admin group. Clients will be associated to a group from the beginning.
- **Xray**: A VPN server providing multiple protocols to bypass censorship: [GitHub Link](https://github.com/XTLS/Xray-core)
- **VPN**: Is an application based on Xray, running on an instance to help clients bypass the censorship.
- **Host** : Is the platform that provides VPS services, for now we only support Vultr.
- **Instance**: Is a VM running on a hosting platform.
- **SSH-Key-Pair**: Is a key pair to access instances. Public-Key will be used to setup instance and ssh key will be used to connect to them and configure them.

### Requirements and Business Invariants

- Clients have to authenticate to the system using basic Auth before being able to interact with the system.
- Besides a username and password, no other data is collected from clients.
- Clients can create a VPN after logging into the system.
- Each client is limited to only one VPN at a time. After logging in, they will see the VPN connection string if they have made one.
- Clients can renew their VPN anytime they wish. Renewing VPNs means deleting the instance and creating a new one.
- Each VPN runs on a single Instance.
- VPNs are not shared among users.

- For each principal, we should store a Default SSH key-pair. This key pair will be used to setup the instance
- SSH public key should be stored in the host, but the private key should be stored in the DB

- Principals can upload multiple startup scripts
- Scripts should be stored in both Host and DB.
- Principal should mark a startup script as default, this will be the scripts that is used to configure instances.

- Principals can upload multiple xray configs.
- Scripts should be stored in DB and uploaded to the instance when initialization is done.

```mermaid
graph TB
    subgraph Access Domain
        subgraph Group aggregate
            Group --has a--> Client
            Group --has a--> Admin
        end
    end
    subgraph Hosting Domain
        User --is associated with a--> Org[Group]
        subgraph group aggregate
            Org --has--> StartUpScript
            Org --has--> SSHKeyPair
            Org --has--> XrayConfig
        end
        User --creates--> Instance
        subgraph instance aggregate
            StartUpScript --used in--> Instance
            SSHKeyPair    --used in--> Instance
            XrayConfig    --used in--> Instance
        end
    end
```

Class diagram for aggregates in the Hosting domain:

```mermaid
classDiagram
    class Instance{
        + UUID ID
        + UUID GroupID
        + UUID ScriptID
        + UUID ConfigID
        + UUID SSHKeyPairID

	    - net.IP IP
	    - Status status
    }

    class Group{
        + UUID ID
        + UUID DefaultSSHKeyPairID
        + UUID DefaultXrayConfigID
        + UUID DefaultStartUpScriptID
        - URL host
        - String APIKey
        + SetDefaultStarUpScript()
        + SetDefaultSSHKeyPair()
        + SetDefaultXrayConfig()
    }

    class SSHKeyPair{
        +UUID ID
    }

    class XrayConfig{
        +UUID ID
    }

    class StartUpScript{
        +UUID ID
    }

    Group "1" --* "*" SSHKeyPair
    Group "1" --* "*" StartUpScript
    Group "1" --* "*" XrayConfig
```

Class diagram for aggregates in the access domain:

```mermaid
classDiagram
    class User{
        + UUID ID
        + String username
        + Hash password
        + UserType type
    }

    class Group {
        + UUID ID
        + InviteClient()
    }

    Group "1" --* "*" User
```
