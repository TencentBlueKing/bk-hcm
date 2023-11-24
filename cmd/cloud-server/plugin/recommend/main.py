import json
import os
import sys
import numpy as np

# problem
from pymoo.core.problem import ElementwiseProblem
# algorithm
from pymoo.algorithms.moo.nsga2 import NSGA2
from pymoo.operators.crossover.sbx import SBX
from pymoo.operators.mutation.pm import PM
from pymoo.operators.sampling.rnd import FloatRandomSampling
# optimize
from pymoo.optimize import minimize
# termination
from pymoo.termination import get_termination


# define problem
class MyProblem(ElementwiseProblem):
    def __init__(self, algorithm_data, n_var, n_obj, n_ieq_constr, xl, xu):
        super().__init__(n_var=n_var,
                         n_obj=n_obj,
                         n_ieq_constr=n_ieq_constr,
                         xl=xl,
                         xu=xu,
                         # type_var=np.int_
                         )
        # ping延迟
        self.PING_INFO = algorithm_data['PING_INFO']
        # 玩家分布
        self.COUNTRY_RATE = algorithm_data['COUNTRY_RATE']
        # 可选IDC列表
        self.IDC_LIST = algorithm_data['IDC_LIST']
        # IDC单位价格
        self.IDC_PRICE = algorithm_data['IDC_PRICE']
        # IDC必选列表
        self.PICK_IDC_LIST = algorithm_data['PICK_IDC_LIST']
        # IDC不选列表
        self.BAN_IDC_LIST = algorithm_data['BAN_IDC_LIST']
        # 覆盖判别最高ping延迟
        self.COVER_PING = algorithm_data['COVER_PING']
        # 玩家覆盖率
        self.COVER_RATE = algorithm_data['COVER_RATE']

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
            ping = sum([
                min([
                    ping for idc, ping in self.PING_INFO[country].items() if idc in idc_list
                ])*rate
                for country, rate in self.COUNTRY_RATE.items()
            ])
        except:
            ping = 100000
        return ping

    def _f2(self, idc_list):
        """IDC单位成本"""
        return sum([price for idc, price in self.IDC_PRICE.items() if idc in idc_list])

    def _g1(self, idc_list):
        """覆盖率计算"""
        try:
            cover_rate = sum([
                rate if
                min([
                    ping for idc, ping in self.PING_INFO[country].items() if idc in idc_list
                ]) <= self.COVER_PING else 0
                for country, rate in self.COUNTRY_RATE.items()
            ])
        except:
            cover_rate = 0.001
        return self.COVER_RATE - cover_rate

    def _get_idc_list(self, x):
        """解码x到IDC"""
        idc_list = [self.IDC_LIST[i] for i, _x in enumerate(x) if _x]
        return list(set(idc_list+self.PICK_IDC_LIST))


# 主函数
def main(plot=False, debug=False):
    """
    主函数
    :return:
    """
    # 1 获取stdin参数
    if debug:
        with open('algorithm_data.json', 'r') as f:
            algorithm_data = json.load(f)
    else:
        algorithm_data = get_stdin()
    # 2 算法实例化
    # 实例化problem
    problem = MyProblem(**{
        "algorithm_data": algorithm_data,
        "n_var": len(algorithm_data['IDC_LIST']),
        "n_obj": 2,
        "n_ieq_constr": 1,
        "xl": np.zeros(len(algorithm_data['IDC_LIST'])),
        "xu": np.ones(len(algorithm_data['IDC_LIST'])),
    })
    # 实例化alogrithm
    algorithm = NSGA2(
        pop_size=40,
        n_offsprings=10,
        sampling=FloatRandomSampling(),
        crossover=SBX(prob=0.9, eta=15),
        mutation=PM(eta=20),
        eliminate_duplicates=True
    )
    # 实例化termination
    termination = get_termination("n_gen", 40)
    # 实例化optimization, 求解
    res = minimize(problem,
                   algorithm,
                   termination,
                   seed=1,
                   save_history=True,
                   # 是否展示迭代详情输出
                   verbose=False)
    # 3 将解映射到实际含义
    pareto = resolve_x(res, algorithm_data)
    if plot:
        plot_pareto_front(res)
    # 4 返回stdout
    sys.stdout.write(json.dumps({"ParetoList": pareto}, ensure_ascii=False) + "\n")
    return True


def resolve_x(res, algorithm_data):
    """
    映射解到实际含义
    :param res:
    :return:
    """
    pareto_list = []
    IDC_LIST = algorithm_data['IDC_LIST']
    PICK_IDC_LIST = algorithm_data['PICK_IDC_LIST']
    COVER_RATE = algorithm_data['COVER_RATE']
    x_list = np.round(res.X)
    for _id, x in enumerate(x_list):
        idc_list = [IDC_LIST[i] for i, _x in enumerate(x) if _x]
        if PICK_IDC_LIST:
            idc_list = list(set(idc_list + PICK_IDC_LIST))
        optimal = {
            'idc': idc_list,
            'f1': f'{res.F[_id][0]:.2f}ms',
            'f2': f'${res.F[_id][1]:.0f}',
            'cover_rate': f'{COVER_RATE - res.G[_id][0]:.2f}'
        }
        if optimal not in pareto_list:
            pareto_list.append(optimal)
    # pareto front排序
    pareto_list = sorted(pareto_list, key=lambda o: float(o['f1'].replace('ms', '')), reverse=False)
    return pareto_list


def plot_pareto_front(res):
    """结果pareto front可视化"""
    import plotly.express as px
    fig = px.scatter(x=res.F[:, 0], y=res.F[:, 1])
    fig.update_xaxes(title='加权ping延迟')
    fig.update_yaxes(title='IDC单位成本')
    fig.update_layout(title='目标函数解空间', title_x=0.5)
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


if __name__ == '__main__':
    main()
