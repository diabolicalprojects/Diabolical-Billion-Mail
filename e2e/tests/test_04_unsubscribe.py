"""E2E: Unsubscribe link → contact deactivated."""

import re
import time
from urllib.parse import parse_qs, urlparse

import httpx
import pytest

from helpers import api
from helpers.wait import poll_until


@pytest.mark.asyncio
async def test_unsubscribe_deactivates_contact(bm_api, mailpit, db, seed_data):
    """Send email → extract unsubscribe JWT → POST to unsub API → verify contact deactivated."""
    await mailpit.clear()

    resp = await api.post(
        bm_api,
        "/api/batch_mail/task/create",
        json={
            "addresser": seed_data["sender_email"],
            "subject": "E2E Unsub Test",
            "full_name": "E2E Sender",
            "template_id": seed_data["template_id"],
            "group_id": seed_data["group_id"],
            "start_time": int(time.time()),
            "track_open": 1,
            "track_click": 1,
            "unsubscribe": 1,
            "threads": 5,
        },
    )
    assert resp.status_code == 200

    msg = await poll_until(
        lambda: mailpit.wait_for_email(
            to=seed_data["recipient_email"],
            subject_contains="E2E Unsub Test",
            timeout=1,
        ),
        description="unsub test email delivery",
        timeout=60,
        interval=2,
    )
    assert msg is not None

    html = await mailpit.get_message_html(msg["ID"])

    # Find unsubscribe link — matches unsubscribe_new.html?jwt=...
    unsub_urls = re.findall(r'href="([^"]*unsubscribe_new\.html\?jwt=[^"]*)"', html, re.IGNORECASE)
    if not unsub_urls:
        # Fallback: broader unsub match
        unsub_urls = re.findall(r'href="([^"]*unsub[^"]*)"', html, re.IGNORECASE)
    assert unsub_urls, "No unsubscribe link found in email"

    # Extract JWT from unsubscribe URL
    parsed = urlparse(unsub_urls[0])
    qs = parse_qs(parsed.query)
    jwt_token = qs.get("jwt", [None])[0]
    assert jwt_token, f"No JWT param in unsubscribe URL: {unsub_urls[0]}"

    # POST directly to unsubscribe API (GET just loads the HTML page with JS)
    base_url = f"{parsed.scheme}://{parsed.netloc}" if parsed.netloc else str(bm_api.base_url)
    async with httpx.AsyncClient(timeout=10) as client:
        unsub_resp = await client.post(
            f"{base_url}/api/unsubscribe_new",
            json={"jwt": jwt_token},
        )
        assert unsub_resp.status_code == 200, f"Unsubscribe POST failed: {unsub_resp.status_code}"

    # Verify contact is deactivated in DB
    row = await db.fetchrow(
        "SELECT active FROM bm_contacts WHERE email = $1 AND group_id = $2",
        seed_data["recipient_email"],
        seed_data["group_id"],
    )
    assert row is not None, "Contact not found in DB after unsubscribe"
    assert row["active"] == 0, "Contact should be deactivated after unsubscribe"
