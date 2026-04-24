import logging
from datetime import date

from apscheduler.schedulers.background import BackgroundScheduler
from fastapi import FastAPI, HTTPException, Query

import aggregator
import config

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)

app = FastAPI(title="Aggregation Service")
scheduler = BackgroundScheduler()


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/recalculate")
def recalculate(target_date: date = Query(default=None)):
    d = target_date or date.today()
    try:
        n = aggregator.run(d)
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc))
    return {"date": str(d), "metrics_written": n}


@app.on_event("startup")
def startup():
    scheduler.add_job(
        aggregator.run_last_days,
        "interval",
        minutes=config.AGGREGATION_INTERVAL_MINUTES,
        id="aggregate",
    )
    scheduler.start()
    log.info("Scheduler started (interval=%dm)", config.AGGREGATION_INTERVAL_MINUTES)


@app.on_event("shutdown")
def shutdown():
    scheduler.shutdown(wait=False)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=config.APP_PORT, log_level="info")
