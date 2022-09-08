import datetime
import json

from helpers.table import table


class Distribution(object):
    def __init__(self, api_name, duration, typ):
        self.api_name = api_name
        self.duration = duration
        self.typ = typ


class PerApiDistribution(object):
    def __init__(self, api_name):
        self.api_name = api_name
        self.sorted_durations = []
        self.max = 0
        self.min = 0
        self.avg = 0
        self.p50 = 0
        self.p75 = 0
        self.p90 = 0
        self.p95 = 0
        self.p99 = 0


class Duration(object):
    def __init__(self, api_name, duration, time):
        self.api_name = api_name
        self.duration = duration
        self.time = time


class Plot(object):
    def __init__(self):
        self.name = ""
        self.iterations = 0
        self.durations = []
        self.distributions = []
        self.ordered_api_names = []

    def append_duration(self, duration):
        if duration is not None:
            self.durations.append(duration)

    def append_zero(self):
        self.durations.append(Duration("", 0, None))

    def calculate_iterations(self):
        self.iterations = 0

        for d in self.durations:
            if d.api_name == "":
                self.iterations += 1

    def api_duration_map(self):
        api_dur_map = {}
        self.ordered_api_names = []

        for d in self.durations:
            if d.api_name == "":
                continue

            if d.api_name not in api_dur_map:
                api_dur_map[d.api_name] = []
                self.ordered_api_names.append(d.api_name)

            api_dur_map[d.api_name].append(d.duration)

        return api_dur_map

    def get_percentile_duration(self, durations, percent):
        return durations[int((percent * len(durations)) / 100)]

    def calc_distributions(self):
        api_dur_map = self.api_duration_map()
        self.distributions = []

        for name in self.ordered_api_names:
            durations = sorted(api_dur_map[name])

            d = PerApiDistribution(name)
            d.sorted_durations = durations
            d.min = durations[0]
            d.max = durations[len(durations) - 1]
            d.avg = int(sum(durations) / len(durations))
            d.p50 = self.get_percentile_duration(durations, 50)
            d.p75 = self.get_percentile_duration(durations, 75)
            d.p90 = self.get_percentile_duration(durations, 90)
            d.p95 = self.get_percentile_duration(durations, 95)
            d.p99 = self.get_percentile_duration(durations, 99)

            self.distributions.append(d)

    def analyze(self, name):
        self.name = name
        self.calculate_iterations()
        self.calc_distributions()

    def to_json(self):
        return json.dumps(
            self,
            default=lambda o: o.microsecond * 1000
            if isinstance(o, datetime.datetime)
            else o.__dict__,
            indent="  ",
        )

    def to_table(self):
        tb = table.Table(
            "Distribution: %s (Iterations %d)" % (self.name, self.iterations),
            ["Dist"] + [d.api_name for d in self.distributions],
            col_width=25,
        )

        tb.append_row(["min"] + [d.min for d in self.distributions])
        tb.append_row(["avg"] + [d.avg for d in self.distributions])
        tb.append_row(["max"] + [d.max for d in self.distributions])
        tb.append_row(["p50"] + [d.p50 for d in self.distributions])
        tb.append_row(["p75"] + [d.p75 for d in self.distributions])
        tb.append_row(["p90"] + [d.p90 for d in self.distributions])
        tb.append_row(["p95"] + [d.p95 for d in self.distributions])
        tb.append_row(["p99"] + [d.p99 for d in self.distributions])

        return str(tb)
