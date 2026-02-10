"""E2E: Contact management — import, groups, tag filtering."""

import json

import pytest

from helpers import api


@pytest.mark.asyncio
async def test_create_group_and_import_contacts(bm_api):
    """Create group → import contacts via paste format → verify count."""
    # Create group
    resp = await api.post(
        bm_api,
        "/api/contact/group/create",
        json={"name": "E2E Contact Test Group", "description": "Testing contact mgmt"},
    )
    assert resp.status_code == 200
    body = resp.json()
    group_id = body.get("data", {}).get("id")
    assert group_id, f"Group creation failed: {body}"

    # Import contacts using correct API contract
    contacts_str = (
        'contact-a@e2e.test,{"first_name":"Alice"}\n'
        'contact-b@e2e.test,{"first_name":"Bob"}\n'
        'contact-c@e2e.test,{"first_name":"Charlie"}'
    )
    resp = await api.post(
        bm_api,
        "/api/contact/group/import",
        json={
            "group_ids": [group_id],
            "contacts": contacts_str,
            "import_type": 2,
            "default_active": 1,
            "status": 1,
        },
    )
    assert resp.status_code == 200

    # List contacts and verify
    resp = await api.get(
        bm_api,
        "/api/contact/list",
        params={"group_id": group_id, "page": 1, "page_size": 50},
    )
    assert resp.status_code == 200
    contacts = resp.json().get("data", {}).get("list", [])
    emails = [c["email"] for c in contacts]
    assert "contact-a@e2e.test" in emails
    assert "contact-b@e2e.test" in emails
    assert "contact-c@e2e.test" in emails

    # Cleanup — delete group
    await api.post(bm_api, "/api/contact/group/delete", json={"id": group_id})


@pytest.mark.asyncio
async def test_list_all_groups(bm_api, seed_data):
    """Verify seed group appears in group list."""
    resp = await api.get(bm_api, "/api/contact/group/all")
    assert resp.status_code == 200
    groups = resp.json().get("data", [])
    group_ids = [g["id"] for g in groups]
    assert seed_data["group_id"] in group_ids


@pytest.mark.asyncio
async def test_tag_filtering_and_or_not(bm_api):
    """Create tags → assign to contacts → verify tag_contact_count with AND/OR/NOT."""
    # Create group + import 3 contacts
    resp = await api.post(
        bm_api,
        "/api/contact/group/create",
        json={"name": "E2E Tag Test Group", "description": "Tag filtering test"},
    )
    assert resp.status_code == 200
    group_id = resp.json().get("data", {}).get("id")
    assert group_id

    contacts_str = (
        'tag-a@e2e.test,{"first_name":"Alice"}\n'
        'tag-b@e2e.test,{"first_name":"Bob"}\n'
        'tag-c@e2e.test,{"first_name":"Charlie"}'
    )
    resp = await api.post(
        bm_api,
        "/api/contact/group/import",
        json={
            "group_ids": [group_id],
            "contacts": contacts_str,
            "import_type": 2,
            "default_active": 1,
            "status": 1,
        },
    )
    assert resp.status_code == 200

    # Get contact IDs
    resp = await api.get(
        bm_api, "/api/contact/list",
        params={"group_id": group_id, "page": 1, "page_size": 50},
    )
    contacts = resp.json().get("data", {}).get("list", [])
    contact_map = {c["email"]: c["id"] for c in contacts}
    id_a = contact_map.get("tag-a@e2e.test")
    id_b = contact_map.get("tag-b@e2e.test")
    id_c = contact_map.get("tag-c@e2e.test")
    assert id_a and id_b and id_c, f"Missing contacts: {contact_map}"

    # Create 2 tags
    resp1 = await api.post(bm_api, "/api/tags/create", json={"name": "e2e-tag-alpha"})
    assert resp1.status_code == 200
    tag1_id = resp1.json().get("data", {}).get("id")
    assert tag1_id, f"Tag1 creation failed: {resp1.json()}"

    resp2 = await api.post(bm_api, "/api/tags/create", json={"name": "e2e-tag-beta"})
    assert resp2.status_code == 200
    tag2_id = resp2.json().get("data", {}).get("id")
    assert tag2_id, f"Tag2 creation failed: {resp2.json()}"

    # Assign tag1 to A+B, tag2 to B+C
    resp = await api.post(
        bm_api, "/api/contact/batch_tags_opt",
        json={"contact_ids": [id_a, id_b], "tag_ids": [tag1_id], "opt_type": "add"},
    )
    assert resp.status_code == 200

    resp = await api.post(
        bm_api, "/api/contact/batch_tags_opt",
        json={"contact_ids": [id_b, id_c], "tag_ids": [tag2_id], "opt_type": "add"},
    )
    assert resp.status_code == 200

    # Verify AND (tag1+tag2) → count=1 (only B has both)
    resp = await api.post(
        bm_api, "/api/contact/group/tag_contact_count",
        json={"group_id": group_id, "tag_ids": [tag1_id, tag2_id], "logic": "AND"},
    )
    assert resp.status_code == 200
    and_count = resp.json().get("data", {}).get("count", -1)
    assert and_count == 1, f"AND count expected 1, got {and_count}"

    # Verify OR (tag1|tag2) → count=3 (A,B,C)
    resp = await api.post(
        bm_api, "/api/contact/group/tag_contact_count",
        json={"group_id": group_id, "tag_ids": [tag1_id, tag2_id], "logic": "OR"},
    )
    assert resp.status_code == 200
    or_count = resp.json().get("data", {}).get("count", -1)
    assert or_count == 3, f"OR count expected 3, got {or_count}"

    # Verify NOT (tag1) → count=1 (only C doesn't have tag1)
    resp = await api.post(
        bm_api, "/api/contact/group/tag_contact_count",
        json={"group_id": group_id, "tag_ids": [tag1_id], "logic": "NOT"},
    )
    assert resp.status_code == 200
    not_count = resp.json().get("data", {}).get("count", -1)
    assert not_count == 1, f"NOT count expected 1, got {not_count}"

    # Cleanup
    await api.post(bm_api, "/api/tags/delete", json={"id": tag1_id})
    await api.post(bm_api, "/api/tags/delete", json={"id": tag2_id})
    await api.post(bm_api, "/api/contact/group/delete", json={"id": group_id})
