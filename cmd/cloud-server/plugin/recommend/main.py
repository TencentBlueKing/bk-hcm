import json
import sys

import numpy as np

# algorithm
from pymoo.algorithms.moo.nsga2 import NSGA2
from pymoo.config import Config

# problem
from pymoo.core.problem import ElementwiseProblem
from pymoo.operators.crossover.sbx import SBX
from pymoo.operators.mutation.pm import PM
from pymoo.operators.sampling.rnd import FloatRandomSampling

# optimize
from pymoo.optimize import minimize

# termination
# from pymoo.termination import get_termination


Config.show_compile_hint = False


# define problem
class MyProblem(ElementwiseProblem):
    def __init__(self, algorithm_data, n_var, n_obj, n_ieq_constr, xl, xu):
        super().__init__(
            n_var=n_var,
            n_obj=n_obj,
            n_ieq_constr=n_ieq_constr,
            xl=xl,
            xu=xu,
            # type_var=np.int_
        )
        # ping延迟
        self.PING_INFO = algorithm_data["PING_INFO"]
        # 玩家分布
        self.COUNTRY_RATE = algorithm_data["COUNTRY_RATE"]
        # 可选IDC列表
        self.IDC_LIST = algorithm_data["IDC_LIST"]
        # IDC单位价格
        self.IDC_PRICE = algorithm_data["IDC_PRICE"]
        # IDC必选列表
        self.PICK_IDC_LIST = algorithm_data["PICK_IDC_LIST"]
        # IDC不选列表
        self.BAN_IDC_LIST = algorithm_data["BAN_IDC_LIST"]
        # 覆盖判别最高ping延迟
        self.COVER_PING = algorithm_data["COVER_PING"]
        # 玩家覆盖率
        self.COVER_RATE = algorithm_data["COVER_RATE"]

    def _evaluate(self, x, out, *args, **kwargs):
        # 解二值化
        x = np.round(x)
        # 解析x到idc_list
        idc_list = self._get_idc_list(x)
        # ping延迟
        f1 = self._f1(idc_list)
        # IDC单位成本
        f2 = self._f2(idc_list)

        # 约束
        g1 = self._g1(idc_list)

        out["F"] = [f1, f2]
        out["G"] = [g1]

    def _f1(self, idc_list):
        """ping延迟计算"""
        try:
            ping = sum(
                min(ping for idc, ping in self.PING_INFO[country].items() if idc in idc_list) * rate
                for country, rate in self.COUNTRY_RATE.items()
            )
        except Exception:
            ping = 100000
        return ping

    def _f2(self, idc_list):
        """IDC单位成本"""
        return sum(price for idc, price in self.IDC_PRICE.items() if idc in idc_list)

    def _g1(self, idc_list):
        """覆盖率计算"""
        try:
            cover_rate = sum(
                rate
                if min(ping for idc, ping in self.PING_INFO[country].items() if idc in idc_list) <= self.COVER_PING
                else 0
                for country, rate in self.COUNTRY_RATE.items()
            )
        except Exception:
            cover_rate = 0.001
        return self.COVER_RATE - cover_rate

    def _get_idc_list(self, x):
        """解码x到IDC"""
        idc_list = [self.IDC_LIST[i] for i, _x in enumerate(x) if _x]
        return list(set(idc_list + self.PICK_IDC_LIST))


# 主函数
def main(plot=False, debug=False):
    """
    主函数
    :return:
    """
    # 1 获取stdin参数
    if debug:
        with open("algorithm_data.json", "r") as f:
            algorithm_data = json.load(f)
    else:
        algorithm_data = get_stdin()
    # 算法输入参数校验
    try:
        input_data_validate(algorithm_data)
    except Exception as e:
        sys.stderr.write(f"input data validate failed, error={e}\n")
        sys.exit(1)
    # COUNTRY_RATE_ORIGIN归一化处理
    COUNTRY_RATE = {}
    total = sum(algorithm_data["COUNTRY_RATE_ORIGIN"].values())
    for country, value in algorithm_data["COUNTRY_RATE_ORIGIN"].items():
        COUNTRY_RATE[country] = value / total
    algorithm_data["COUNTRY_RATE"] = COUNTRY_RATE
    # 2 算法实例化
    # 实例化problem
    problem = MyProblem(
        **{
            "algorithm_data": algorithm_data,
            "n_var": len(algorithm_data["IDC_LIST"]),
            "n_obj": 2,
            "n_ieq_constr": 1,
            "xl": np.zeros(len(algorithm_data["IDC_LIST"])),
            "xu": np.ones(len(algorithm_data["IDC_LIST"])),
        }
    )
    # 实例化alogrithm
    algorithm = NSGA2(
        pop_size=40,
        n_offsprings=10,
        sampling=FloatRandomSampling(),
        crossover=SBX(prob=0.9, eta=15),
        mutation=PM(eta=20),
        eliminate_duplicates=True,
    )
    # 实例化termination
    # termination = get_termination("n_gen", 40)
    # 实例化optimization, 求解
    res = minimize(
        problem,
        algorithm,
        ("n_gen", 40),
        seed=1,
        save_history=True,
        # 是否展示迭代详情输出
        verbose=False,
    )
    # 3 将解映射到实际含义
    pareto = resolve_x(res, algorithm_data)
    if plot:
        plot_pareto_front(res)
    # 4 返回stdout
    sys.stdout.write(json.dumps({"PARETO_LIST": pareto}, ensure_ascii=False) + "\n")
    return True


def resolve_x(res, algorithm_data):
    """
    映射解到实际含义
    :param res:
    :return:
    """
    pareto_list = []
    IDC_LIST = algorithm_data["IDC_LIST"]
    PICK_IDC_LIST = algorithm_data["PICK_IDC_LIST"]
    COVER_RATE = algorithm_data["COVER_RATE"]
    COVER_PING_RANGES = algorithm_data.get("COVER_PING_RANGES", [])
    IDC_PRICE_RANGES = algorithm_data.get("IDC_PRICE_RANGES", [])
    x_list = np.round(res.X)
    for _id, x in enumerate(x_list):
        idc_list = [IDC_LIST[i] for i, _x in enumerate(x) if _x]
        if PICK_IDC_LIST:
            idc_list = list(set(idc_list + PICK_IDC_LIST))
        optimal = {
            "IDC": idc_list,
            "F1": res.F[_id][0],
            "F2": res.F[_id][1],
            "COVER_RATE": COVER_RATE - res.G[_id][0],
            # 新增解析f1与f2的评分
            "F1_SCORE": get_score_for_function(COVER_PING_RANGES, res.F[_id][0]),
            "F2_SCORE": get_score_for_function(IDC_PRICE_RANGES, res.F[_id][1]),
        }
        if optimal not in pareto_list and idc_list:
            pareto_list.append(optimal)
    # pareto front排序
    pareto_list = sorted(pareto_list, key=lambda o: o["F1"], reverse=False)
    return pareto_list


def get_score_for_function(score_list, f_value):
    """
    将函数输出按照分数区间转换成打分
    :param score_list:
    :param f_value:
    :return:
    """
    for s in score_list:
        if s["range"][0] <= f_value < s["range"][1]:
            return s["score"]
    return -1


def plot_pareto_front(res):
    """结果pareto front可视化"""
    import plotly.express as px

    fig = px.scatter(x=res.F[:, 0], y=res.F[:, 1])
    fig.update_xaxes(title="加权ping延迟")
    fig.update_yaxes(title="IDC单位成本")
    fig.update_layout(title="目标函数解空间", title_x=0.5)
    fig.show()
    return True


def get_stdin():
    """
    从stdin中获取配置参数
    :return:
    """
    try:
        ALGORITHM_INPUT_DATA = json.load(sys.stdin)
    except json.JSONDecodeError:
        sys.stderr.write("Error: Invalid JSON\n")
        sys.exit(1)
    return ALGORITHM_INPUT_DATA


def input_data_validate(algorithm_data):
    """
    算法入参参数校验
    :return:
    """
    # 1 检查必须参数
    required_fields = [
        "COUNTRY_RATE_ORIGIN",
        "COVER_RATE",
        "COVER_PING",
        "PING_INFO",
        "IDC_PRICE",
        "IDC_LIST",
        "COVER_PING_RANGES",
        "IDC_PRICE_RANGES",
        "BAN_IDC_LIST",
        "PICK_IDC_LIST",
    ]
    missing_fields = set(required_fields) - set(algorithm_data.keys())
    if missing_fields:
        raise ValueError(f"Missing required fields: {', '.join(missing_fields)}")

    # 2 检查参数类型和值
    check_types = {
        "COUNTRY_RATE_ORIGIN": (dict, (float, int)),
        "COVER_RATE": (float,),
        "COVER_PING": ((int, float),),
        "PING_INFO": (dict, dict, (float, int)),
        "IDC_PRICE": (dict, (float, int)),
        "IDC_LIST": (list,),
        "COVER_PING_RANGES": (list,),
        "IDC_PRICE_RANGES": (list,),
        "BAN_IDC_LIST": (list,),
        "PICK_IDC_LIST": (list,),
    }

    for key, types in check_types.items():
        value = algorithm_data.get(key)
        # first level
        if not isinstance(value, types[0]):
            raise TypeError(f"{key} must be {types[0].__name__}")
        # second level
        if isinstance(value, dict):
            for sub_key, sub_value in value.items():
                if not isinstance(sub_value, dict):
                    if not sub_key:
                        raise TypeError(f"{key}/{sub_key} key can not be null")
                    if not isinstance(sub_value, types[1]):
                        raise TypeError(
                            f"{key}/{sub_key} value must " f"be {types[1][0].__name__} or {types[1][1].__name__}"
                        )
                # third level
                else:
                    for sub_sub_key, sub_sub_value in sub_value.items():
                        if not sub_sub_key:
                            raise TypeError(f"{key}/{sub_key}/{sub_sub_key} key can not be null")
                        if not isinstance(sub_sub_value, types[2]):
                            raise TypeError(
                                f"{key}/{sub_key}/{sub_sub_key} value must "
                                f"be {types[2][0].__name__} or {types[2][1].__name__}"
                            )
        if key in ["COVER_PING_RANGES", "IDC_PRICE_RANGES"]:
            for item in value:
                if "score" not in item or "range" not in item:
                    raise KeyError(f"{key} must contain score, range")
                if not isinstance(item["score"], (int, float)):
                    raise ValueError(f"{key} score must be int or float")
                if not isinstance(item["range"], list):
                    raise ValueError(f"{key} score must be list")
                for range_value in item["range"]:
                    if not isinstance(range_value, (int, float)):
                        raise ValueError(f'{key}/{item["range"]}/{range_value} range value must be int or float')

    # 3 检查键值对应关系
    if not set(algorithm_data["COUNTRY_RATE_ORIGIN"].keys()).issubset(set(algorithm_data["PING_INFO"].keys())):
        missing_keys = set(algorithm_data["COUNTRY_RATE_ORIGIN"].keys()) - set(algorithm_data["PING_INFO"].keys())
        raise KeyError(f"COUNTRY_RATE_ORIGIN and PING_INFO has different keys: {missing_keys}")

    if not set(algorithm_data["IDC_LIST"]).issubset(set(algorithm_data["IDC_PRICE"].keys())):
        missing_keys = set(algorithm_data["IDC_LIST"]) - set(algorithm_data["IDC_PRICE"].keys())
        raise KeyError(f"IDC_LIST and PING_INFO has different keys: {missing_keys}")

    return True


if __name__ == "__main__":
    main()