from collections.abc import Iterator

class Sequence(Iterator):

    def __init__(self, initial=1, step=1, max_value=None):
        self._counter = 0
        self._initial = initial
        self._step = step
        self._max = max_value

    def __next__(self):
        tmp_result = self._initial + self._counter
        self._counter += self._step
        if self._max is not None and tmp_result > self._max:
            self.reset()
            self._counter += self._step
            return self._initial
        return tmp_result

    def reset(self):
        self._counter = 0


seq_name = Sequence()
seq_pass = Sequence()


def get_name() -> str:
    return str(next(seq_name))


def get_pass() -> str:
    return str(next(seq_pass))
