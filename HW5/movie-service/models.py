from enum import Enum
from pydantic import BaseModel, field_validator


class EventType(str, Enum):
    VIEW_STARTED  = "VIEW_STARTED"
    VIEW_FINISHED = "VIEW_FINISHED"
    VIEW_PAUSED   = "VIEW_PAUSED"
    VIEW_RESUMED  = "VIEW_RESUMED"
    LIKED         = "LIKED"
    SEARCHED      = "SEARCHED"


class DeviceType(str, Enum):
    MOBILE  = "MOBILE"
    DESKTOP = "DESKTOP"
    TV      = "TV"
    TABLET  = "TABLET"


class MovieEvent(BaseModel):
    event_id:         str = ""
    user_id:          str
    movie_id:         str
    event_type:       EventType
    timestamp:        str = ""
    device_type:      DeviceType
    session_id:       str
    progress_seconds: int = 0

    @field_validator("progress_seconds")
    @classmethod
    def non_negative(cls, v: int) -> int:
        if v < 0:
            raise ValueError("progress_seconds must be >= 0")
        return v
