"""
@cgraph-id py-root
@cgraph-deps sql-root
Module-level tasks for ETL.
"""


def handle(event):
    print("processing", event)
    return {"ok": True}
