# Generated by the protocol buffer compiler.  DO NOT EDIT!
# sources: localaccount/v1/localaccount.proto
# plugin: python-betterproto
from dataclasses import dataclass

import betterproto


@dataclass
class LocalAccountResource(betterproto.Message):
    # resource identifier
    resource_id: str = betterproto.string_field(1)
    # Username provided by admin
    username: str = betterproto.string_field(2)
    # SSH Public Key of EN
    ssh_key: str = betterproto.string_field(3)
    # Tenant Identifier.
    tenant_id: str = betterproto.string_field(100)
    # Creation timestamp
    created_at: str = betterproto.string_field(200)
    updated_at: str = betterproto.string_field(201)
