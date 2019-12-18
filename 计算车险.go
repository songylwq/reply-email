package main

import (
	"fmt"
)

type Plan struct {
	//剩余本金
	Benjin float32
	//刷卡费
	Shuakafei float32
}

func main() {
	//贷款金额
	var benjin float32 = 75000
	//还款月数
	var monthNum float32 = 4 * 12
	//每月平均本金
	var monthBenjin = benjin / monthNum
	//信用卡利率
	var rate float32 = 0.002

	//还款计划切片，剩余本金
	var planSlice = make([]Plan, int(monthNum), int(monthNum))

	//累计刷卡费
	var leijiFei float32
	var leijiShuaka float32
	for ind,p := range planSlice{
		p.Benjin = benjin - float32(ind) * monthBenjin
		p.Shuakafei = p.Benjin*rate
		leijiFei += p.Shuakafei
		leijiShuaka += p.Benjin
		fmt.Printf("第 %v 个月,还款本金:%.2f, 本金剩余：%.2f, 本次刷卡费 %.2f, 累计刷卡费 %.2f , 累计刷卡额：%.2f万; \n",
			ind,
			monthBenjin,
			p.Benjin,
			p.Shuakafei,
			leijiFei,
			leijiShuaka/10000	)
	}

	fmt.Printf("本金:%.2f, 总刷卡费：%.2f, 总共还款 %.2f\n",
		benjin,
		leijiFei,
		benjin + leijiFei	)
}
