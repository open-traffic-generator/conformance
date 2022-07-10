class Duration(object):
    def __init__(self, api_name, duration, time):
        self.api_name = api_name
        self.duration = duration
        self.time = time


class Plot(object):
    def __init__(self):
        self.durations = []

    def append_duration(self, duration):
        if duration is not None:
            self.durations.append(duration)

    def append_zero(self):
        self.durations.append(Duration(None, 0, None))
