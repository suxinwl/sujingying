<template>
  <div class="funds-page">
    <van-nav-bar
      title="资金"
      fixed
      placeholder
    />
    
    <!-- 资金概览 -->
    <div class="balance-card">
      <div class="balance-item">
        <div class="label">总定金</div>
        <div class="amount">¥{{ formatMoney(totalDeposit) }}</div>
      </div>
      <div class="balance-row">
        <div class="balance-item">
          <div class="label">可用定金</div>
          <div class="amount">¥{{ formatMoney(userInfo.available_deposit) }}</div>
        </div>
        <div class="balance-item">
          <div class="label">持单定金</div>
          <div class="amount">¥{{ formatMoney(holdingMargin) }}</div>
        </div>
        <div class="balance-item">
          <div class="label">待退定金</div>
          <div class="amount">¥{{ formatMoney(pendingRefundDeposit) }}</div>
        </div>
      </div>
      
      <div class="actions">
        <van-button type="primary" size="small" @click="showDeposit = true">
          付定金
        </van-button>
        <van-button plain type="primary" size="small" @click="showWithdraw = true">
          退定金
        </van-button>
      </div>
    </div>
    
    <!-- 资金流水 -->
    <van-tabs v-model:active="activeTab" @change="loadRecords">
      <van-tab title="全部" name="all" />
      <van-tab title="付定金" name="deposit" />
      <van-tab title="退定金" name="withdraw" />
      <van-tab title="补定金" name="trade" />
    </van-tabs>
    
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadRecords"
      >
        <template v-if="activeTab === 'trade'">
          <div class="trade-filters">
            <div class="trade-type-tabs">
              <div
                class="trade-type-tab trade-type-long"
                :class="{ active: tradeType === 'long_buy' }"
                @click="tradeType = 'long_buy'"
              >
                锁价买料
              </div>
              <div
                class="trade-type-tab trade-type-short"
                :class="{ active: tradeType === 'short_sell' }"
                @click="tradeType = 'short_sell'"
              >
                锁价卖料
              </div>
            </div>
          </div>
        </template>

        <div v-if="visibleRecords.length === 0" class="empty">
          <van-empty :description="activeTab === 'trade' ? '暂无补定金记录' : '暂无记录'" />
        </div>
        
        <div 
          v-for="record in visibleRecords" 
          :key="record.id" 
          class="record-card"
          @click="showRecordDetail(record)"
        >
          <div class="record-card-header">
            <div
              class="record-type-badge"
              :class="[
                getRecordTypeClass(record.type),
                record.type === 'trade'
                  ? (record.order?.type === 'long_buy'
                      ? 'type-trade-long'
                      : (record.order?.type === 'short_sell' ? 'type-trade-short' : ''))
                  : ''
              ]"
            >
              {{
                record.type === 'trade'
                  ? '补定金-' + getOrderTypeText(record.order?.type)
                  : getRecordTypeText(record.type)
              }}
            </div>
            <div class="record-status" :class="record.status">
              {{ getStatusText(record.status) }}
            </div>
          </div>
          
          <div class="record-card-body">
            <div class="record-info-row">
              <span class="label">订单号</span>
              <span class="value">#{{ record.id }}</span>
            </div>
            <div class="record-info-row" v-if="record.type === 'trade'">
              <span class="label">订单类型</span>
              <span class="value">{{ getOrderTypeText(record.order?.type) }}</span>
            </div>
            <div class="record-info-row" v-if="record.type === 'deposit'">
              <span class="label">充值前可用定金</span>
              <span class="value">¥{{ formatMoney(record.before_balance || 0) }}</span>
            </div>
            <div class="record-info-row">
              <span class="label">
                {{ record.type === 'deposit' ? '充值金额' : (record.type === 'trade' ? '定金' : '金额') }}
              </span>
              <span
                class="value amount-text"
                :class="{
                  income: getRecordAmountDisplay(record) > 0,
                  expense: getRecordAmountDisplay(record) < 0
                }"
              >
                {{
                  getRecordAmountDisplay(record) > 0
                    ? '+'
                    : getRecordAmountDisplay(record) < 0
                      ? '-'
                      : ''
                }}¥{{ formatMoney(Math.abs(getRecordAmountDisplay(record))) }}
              </span>
            </div>
            <div
              class="record-info-row"
              v-if="record.type === 'trade' && record.order && getSupplementDepositDisplay(record.order) > 0"
            >
              <span class="label">已补定金</span>
              <span class="value">¥{{ formatMoney(getSupplementDepositDisplay(record.order)) }}</span>
            </div>
            <div class="record-info-row" v-if="record.type === 'trade' && record.order">
              <span class="label">{{ isOrderHolding(record.order) ? '浮动盈亏' : '结算盈亏' }}</span>
              <span
                class="value amount-text"
                :class="{
                  income: getTradeRecordPnl(record) > 0,
                  expense: getTradeRecordPnl(record) < 0
                }"
              >
                {{
                  getTradeRecordPnl(record) > 0
                    ? '+'
                    : getTradeRecordPnl(record) < 0
                      ? '-'
                      : ''
                }}¥{{ formatMoney(Math.abs(getTradeRecordPnl(record))) }}
              </span>
            </div>
            <div class="record-info-row" v-if="record.type === 'trade' && record.order">
              <span class="label">需补定金</span>
              <span class="value">
                ¥{{ formatMoney(calcOrderNeedSupplement(record.order)) }}
              </span>
            </div>
            <div class="record-info-row" v-if="record.type === 'deposit'">
              <span class="label">充值后可用定金</span>
              <span class="value">¥{{ formatMoney(record.after_balance || 0) }}</span>
            </div>
            <div class="record-info-row">
              <span class="label">{{ record.type === 'deposit' ? '充值日期' : '日期' }}</span>
              <span class="value">{{ formatDateTime(record.created_at) }}</span>
            </div>
            <div class="record-info-row" v-if="record.description">
              <span class="label">备注</span>
              <span class="value desc-text">{{ record.description }}</span>
            </div>
          </div>
          
          <div class="record-card-footer">
            <div
              v-if="record.type === 'trade' && record.order"
              class="record-actions"
            >
              <van-button
                v-if="calcOrderNeedSupplement(record.order) > 0"
                size="small"
                type="warning"
                plain
                @click.stop="onClickSupplement(record.order)"
              >
                补定金
              </van-button>
              <van-button
                v-if="isOrderHolding(record.order)"
                size="small"
                type="primary"
                @click.stop="openSettleDialog(record.order)"
              >
                结算
              </van-button>
              <van-button
                size="small"
                type="default"
                @click.stop="openTradeDetail(record.order)"
              >
                查看料单
              </van-button>
            </div>
            <div
              v-else-if="(['supplement_deposit', 'supplement'].includes((record.type || '').toLowerCase())) && record.order_id"
              class="record-actions"
            >
              <van-button
                size="small"
                type="default"
                @click.stop="openTradeDetail(record)"
              >
                查看料单
              </van-button>
            </div>
            <span class="view-detail">查看详情 ></span>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
    
    <!-- 付定金弹窗 -->
    <van-popup v-model:show="showDeposit" position="bottom" round :style="{ height: '90%' }">
      <div class="deposit-popup">
        <van-nav-bar
          title="钱包付定金"
          left-arrow
          @click-left="showDeposit = false"
        />
        
        <div class="deposit-content">
          <!-- 金额输入 -->
          <div class="amount-section">
            <div class="amount-label">金额</div>
            <van-field
              v-model="depositForm.amount"
              type="digit"
              placeholder="请输入金额"
              class="amount-input"
            />
          </div>
          
          <!-- 快捷金额 -->
          <div class="quick-amounts">
            <van-button 
              v-for="amount in quickAmounts" 
              :key="amount"
              size="small"
              :type="depositForm.amount == amount ? 'primary' : 'default'"
              @click="depositForm.amount = amount"
            >
              {{ formatQuickAmount(amount) }}
            </van-button>
          </div>
          
          <!-- 付款账户 -->
          <div class="section">
            <div class="section-title">付款账户</div>
            <van-cell
              title="选择付款账户"
              is-link
              :value="selectedPaymentCard ? selectedPaymentCard.bank_name : ''"
              @click="showPaymentCardPicker = true"
            />
            <div v-if="selectedPaymentCard" class="card-info">
              <div class="info-row">
                <span class="label">账户类型</span>
                <span class="value">银行卡</span>
              </div>
              <div class="info-row">
                <span class="label">户名</span>
                <span class="value">{{ selectedPaymentCard.card_holder }}</span>
              </div>
              <div class="info-row">
                <span class="label">产名</span>
                <span class="value">{{ selectedPaymentCard.bank_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">账户</span>
                <span class="value">{{ selectedPaymentCard.card_number }}</span>
              </div>
            </div>
          </div>
          
          <!-- 收款账户 -->
          <div class="section">
            <div class="section-title">收款账户</div>
            <div class="tip-text">点击删除或者点击设置为默认</div>
            <div v-if="paymentInfo.bank_card" class="card-info" style="margin-bottom: 12px;">
              <div class="info-row">
                <span class="label">账户类型</span>
                <span class="value">银行卡</span>
              </div>
              <div class="info-row">
                <span class="label">户名</span>
                <span class="value">{{ paymentInfo.bank_card.account_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">产名</span>
                <span class="value">{{ paymentInfo.bank_card.bank_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">银行</span>
                <span class="value">{{ paymentInfo.bank_card.branch_name || paymentInfo.bank_card.bank_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">账户</span>
                <span class="value">{{ paymentInfo.bank_card.account_number }}</span>
              </div>
            </div>
          </div>
          
          <!-- 支付凭证 -->
          <div class="section">
            <div class="section-title">支付凭证</div>
            <van-uploader
              v-model="voucherFiles"
              :max-count="5"
              :after-read="afterReadVoucher"
              multiple
            />
          </div>
          
          <!-- 温馨提示 -->
          <div class="section">
            <div class="section-title">温馨提示</div>
            <div class="tip-content">
              为保证资金安全请到任何机构或便利店使用现金（不和银卡转的是所有银行支付账单会发给对方支付支口账号）特10001-11本1000,11本-1000）方便快速认证系统自动打款及对账安全请注意选择正确
            </div>
          </div>
          
          <!-- 备注 -->
          <div class="section">
            <div class="section-title">备注</div>
            <van-field
              v-model="depositForm.note"
              type="textarea"
              placeholder="请输入内容"
              rows="2"
            />
          </div>
          
          <!-- 协议 -->
          <div class="agreement">
            <van-checkbox v-model="agreeProtocol">
              请仔细阅读并同意
              <span style="color: #ee0a24;">账户打款者姓名需一致协议</span>
            </van-checkbox>
          </div>
          
          <!-- 提交按钮 -->
          <div class="submit-btn">
            <van-button
              round
              block
              type="danger"
              @click="onDeposit"
              :disabled="!agreeProtocol"
            >
              提交审核
            </van-button>
          </div>
        </div>
      </div>
    </van-popup>
    
    <!-- 付款账户选择器 -->
    <van-popup v-model:show="showPaymentCardPicker" position="bottom" round>
      <van-picker
        :columns="paymentCardColumns"
        @confirm="onSelectPaymentCard"
        @cancel="showPaymentCardPicker = false"
      />
    </van-popup>
    
    <!-- 退定金弹窗 -->
    <van-popup v-model:show="showWithdraw" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>退定金</h3>
        </div>
        <van-form @submit="onWithdraw">
          <van-field
            v-model="withdrawForm.amount"
            type="number"
            label="退定金金额"
            placeholder="请输入退定金金额"
            :rules="[{ required: true, message: '请输入退定金金额' }]"
          >
            <template #extra>
              <span style="color: #999; font-size: 12px;">
                可用: ¥{{ formatMoney(userInfo.available_deposit) }}
              </span>
            </template>
          </van-field>
          <van-field
            v-model="selectedBankCardText"
            label="银行卡"
            placeholder="请选择银行卡"
            readonly
            is-link
            @click="openBankCardPicker('withdraw')"
            :rules="[{ required: true, message: '请选择银行卡' }]"
          />
	      <van-field
	        v-model="withdrawForm.note"
	        type="textarea"
	        label="备注"
	        placeholder="可填写提现说明（选填）"
	        rows="2"
	      />
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              确认退定金
            </van-button>
          </div>
        </van-form>
      </div>
    </van-popup>
    
    <!-- 银行卡选择 -->
    <van-action-sheet v-model:show="showBankCardPicker" title="选择银行卡">
      <div class="bank-card-list">
        <div
          v-for="card in bankCards"
          :key="card.id || card.ID"
          class="bank-card-item"
          @click="selectBankCard(card)"
        >
          <div class="card-info">
            <div class="bank-name">{{ card.bank_name || card.BankName }}</div>
            <div class="card-number">**** **** **** {{ (card.card_number || card.CardNumber || '').slice(-4) }}</div>
          </div>
          <van-icon name="success" v-if="card.is_default || card.IsDefault" color="#07c160" />
        </div>
        <div v-if="bankCards.length === 0" class="empty-tip">
          暂无银行卡，<span style="color: #1989fa; cursor: pointer;" @click="goToAddCard">点击添加</span>
        </div>
      </div>
    </van-action-sheet>
    
    <!-- 详情弹窗 -->
    <van-popup v-model:show="showDepositDetailDialog" position="bottom" round :style="{ height: '80%' }">
      <div class="detail-popup" v-if="currentDetailRecord">
        <van-nav-bar
          :title="getRecordTypeText(currentDetailRecord.type) + '详情'"
          left-arrow
          @click-left="showDepositDetailDialog = false"
        />
        
        <div class="detail-content">
          <van-cell-group inset>
            <!-- 锁价交易详情 -->
            <template v-if="currentDetailRecord.type === 'trade'">
              <van-cell
                title="订单号"
                :value="'#' + (currentDetailRecord.order?.order_id || currentDetailRecord.id)"
              />
              <van-cell
                title="订单类型"
                :value="getOrderTypeText(currentDetailRecord.order?.type)"
              />
              <van-cell
                title="订单状态"
                :value="getStatusText(currentDetailRecord.status || currentDetailRecord.order?.status)"
              />
              <van-cell
                title="锁定单价"
                :value="'¥' + formatMoney(currentDetailRecord.order?.locked_price) + ' /克'"
              />
              <van-cell
                title="锁定货款"
                :value="'¥' + formatMoney((currentDetailRecord.order?.locked_price || 0) * (currentDetailRecord.order?.weight_g || 0))"
              />
              <van-cell
                title="当前价格"
                :value="'¥' + formatMoney(calcOrderCurrentPrice(currentDetailRecord.order)) + ' /克'"
              />
              <van-cell
                title="克重"
                :value="(currentDetailRecord.order?.weight_g || 0) + ' 克'"
              />
              <van-cell
                title="定金"
                :value="'¥' + formatMoney(getBaseDeposit(currentDetailRecord.order))"
              />
              <van-cell
                title="已补定金"
                :value="'¥' + formatMoney(getSupplementDepositDisplay(currentDetailRecord.order))"
              />
              <van-cell
                :title="isOrderHolding(currentDetailRecord.order) ? '浮动盈亏' : '结算盈亏'"
                :value="
                  (getOrderDisplayPnl(currentDetailRecord.order) > 0
                    ? '+'
                    : getOrderDisplayPnl(currentDetailRecord.order) < 0
                      ? '-'
                      : '') +
                  '¥' +
                  formatMoney(Math.abs(getOrderDisplayPnl(currentDetailRecord.order)))
                "
                :class="{
                  'pnl-profit-cell': getOrderDisplayPnl(currentDetailRecord.order) > 0,
                  'pnl-loss-cell': getOrderDisplayPnl(currentDetailRecord.order) < 0
                }"
              />
              <van-cell
                title="需补定金"
                :value="'¥' + formatMoney(calcOrderNeedSupplement(currentDetailRecord.order))"
              />
              <van-cell
                title="定金率"
                :value="calcOrderMarginRate(currentDetailRecord.order) !== null ? calcOrderMarginRate(currentDetailRecord.order).toFixed(2) + '%' : '-'"
              />
              <van-cell
                title="创建时间"
                :value="formatDateTime(currentDetailRecord.order?.created_at || currentDetailRecord.created_at)"
              />
            </template>

            <!-- 充值/提现明细 -->
            <template v-else>
              <van-cell title="订单号" :value="'#' + currentDetailRecord.id" />
              <van-cell title="类型" :value="getRecordTypeText(currentDetailRecord.type)" />
              <van-cell 
                v-if="currentDetailRecord.status" 
                title="状态" 
                :value="getStatusText(currentDetailRecord.status)" 
                :label-class="currentDetailRecord.status"
              />
              <van-cell 
                title="金额" 
                :value="(currentDetailRecord.amount > 0 ? '+' : '') + '¥' + formatMoney(Math.abs(currentDetailRecord.amount))"
                :class="{ 'income-cell': currentDetailRecord.amount > 0, 'expense-cell': currentDetailRecord.amount < 0 }"
              />
              <van-cell 
                v-if="currentDetailRecord.before_balance !== undefined" 
                title="变动前余额" 
                :value="'¥' + formatMoney(currentDetailRecord.before_balance)" 
              />
              <van-cell 
                v-if="currentDetailRecord.after_balance !== undefined" 
                title="变动后余额" 
                :value="'¥' + formatMoney(currentDetailRecord.after_balance)" 
              />
              <van-cell 
                v-if="currentDetailRecord.method" 
                title="支付方式" 
                :value="currentDetailRecord.method === 'bank' ? '银行转账' : currentDetailRecord.method" 
              />
              <van-cell title="时间" :value="formatDateTime(currentDetailRecord.created_at)" />
              <van-cell 
                v-if="currentDetailRecord.reviewed_at" 
                title="审核时间" 
                :value="formatDateTime(currentDetailRecord.reviewed_at)" 
              />
              <van-cell 
                v-if="currentDetailRecord.paid_at" 
                title="打款时间" 
                :value="formatDateTime(currentDetailRecord.paid_at)" 
              />
              <van-cell 
                v-if="currentDetailRecord.description" 
                title="备注" 
                :value="currentDetailRecord.description" 
              />
            </template>
          </van-cell-group>
          <div
            v-if="currentDetailRecord.type === 'trade' && currentDetailRecord.order"
            class="detail-actions"
          >
            <van-button
              v-if="calcOrderNeedSupplement(currentDetailRecord.order) > 0"
              type="warning"
              block
              size="small"
              @click="onClickSupplement(currentDetailRecord.order)"
            >
              补定金
            </van-button>
            <van-button
              v-if="isOrderHolding(currentDetailRecord.order)"
              type="primary"
              block
              size="small"
              style="margin-top: 8px;"
              @click="openSettleDialog(currentDetailRecord.order)"
            >
              结算
            </van-button>
            <van-button
              type="default"
              block
              size="small"
              style="margin-top: 8px;"
              @click="openTradeDetail(currentDetailRecord.order)"
            >
              查看料单
            </van-button>
          </div>
          
          <!-- 支付凭证 -->
          <div v-if="currentDetailRecord.voucher_url" class="voucher-section">
            <div class="section-title">支付凭证</div>
            <div class="voucher-images">
              <van-image
                v-for="(url, index) in getVoucherUrls(currentDetailRecord.voucher_url)"
                :key="index"
                :src="url"
                width="100"
                height="100"
                fit="cover"
                @click="previewVoucher(currentDetailRecord.voucher_url, index)"
              />
            </div>
          </div>
        </div>
      </div>
    </van-popup>

    <!-- 锁价订单详情弹窗（查看料单） -->
    <van-popup v-model:show="showTradeDetailDialog" position="bottom" round :style="{ height: '80%' }">
      <div class="detail-popup" v-if="currentDetailOrder">
        <van-nav-bar
          :title="getOrderTypeText(currentDetailOrder.type) + '订单详情'"
          left-arrow
          @click-left="showTradeDetailDialog = false"
        />

        <div class="detail-content">
          <van-cell-group inset>
            <van-cell
              title="订单号"
              :value="'#' + (currentDetailOrder.order_id || currentDetailOrder.id)"
            />
            <van-cell
              title="订单类型"
              :value="getOrderTypeText(currentDetailOrder.type)"
            />
            <van-cell
              title="订单状态"
              :value="getStatusText(currentDetailOrder.status)"
            />
            <van-cell
              title="锁定单价"
              :value="'¥' + formatMoney(currentDetailOrder.locked_price || 0) + ' /克'"
            />
            <van-cell
              title="锁定货款"
              :value="'¥' + formatMoney((currentDetailOrder.locked_price || 0) * (currentDetailOrder.weight_g || 0))"
            />
            <van-cell
              title="当前价格"
              :value="'¥' + formatMoney(calcOrderCurrentPrice(currentDetailOrder)) + ' /克'"
            />
            <van-cell
              title="克重"
              :value="(currentDetailOrder.weight_g || 0) + ' 克'"
            />
            <van-cell
              title="定金"
              :value="'¥' + formatMoney(getBaseDeposit(currentDetailOrder))"
            />
            <van-cell
              title="已补定金"
              :value="'¥' + formatMoney(getSupplementDepositDisplay(currentDetailOrder))"
            />
            <van-cell
              :title="isOrderHolding(currentDetailOrder) ? '浮动盈亏' : '结算盈亏'"
              :value="
                (getOrderDisplayPnl(currentDetailOrder) > 0
                  ? '+'
                  : getOrderDisplayPnl(currentDetailOrder) < 0
                    ? '-'
                    : '') +
                '¥' +
                formatMoney(Math.abs(getOrderDisplayPnl(currentDetailOrder)))
              "
              :class="{
                'pnl-profit-cell': getOrderDisplayPnl(currentDetailOrder) > 0,
                'pnl-loss-cell': getOrderDisplayPnl(currentDetailOrder) < 0
              }"
            />
            <van-cell
              title="需补定金"
              :value="'¥' + formatMoney(calcOrderNeedSupplement(currentDetailOrder))"
            />
            <van-cell
              title="定金率"
              :value="
                calcOrderMarginRate(currentDetailOrder) !== null
                  ? calcOrderMarginRate(currentDetailOrder).toFixed(2) + '%'
                  : '-'
              "
            />
            <van-cell
              title="创建时间"
              :value="formatDateTime(currentDetailOrder.created_at)"
            />
          </van-cell-group>
        </div>
      </div>
    </van-popup>

    <!-- 订单结算弹窗 -->
    <van-dialog
      v-model:show="showSettleDialog"
      title="结算订单"
      show-cancel-button
      @confirm="confirmSettle"
    >
      <van-field
        v-model="settlePayPassword"
        label="支付密码"
        type="password"
        placeholder="请输入支付密码"
      />
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { showToast, showDialog, showImagePreview } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney, formatDateTime } from '../utils/helpers'
import { useQuoteStore } from '../stores/quote'

const activeTab = ref('all')
const records = ref([])
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)
const page = ref(1)

const userInfo = ref({
  available_deposit: 0,
  used_deposit: 0
})

const pendingRefundDeposit = ref(0)

const quoteStore = useQuoteStore()
const holdingOrders = ref([])

const calcOrderCurrentPrice = (order) => {
  if (!order) return 0
  if (order.type === 'long_buy') {
    return quoteStore.sellPrice || quoteStore.buyPrice || order.locked_price || 0
  }
  if (order.type === 'short_sell') {
    return quoteStore.buyPrice || quoteStore.sellPrice || order.locked_price || 0
  }
  return order.locked_price || 0
}

const calcOrderPnl = (order) => {
  if (!order) return 0
  const price = calcOrderCurrentPrice(order)
  const locked = order.locked_price || 0
  const weight = order.weight_g || 0
  if (!weight) return 0
  if (order.type === 'long_buy') {
    return (price - locked) * weight
  }
  if (order.type === 'short_sell') {
    return (locked - price) * weight
  }
  return 0
}

// 基础定金：按业务约定为 交易克重 × 10 元
const getBaseDeposit = (order) => {
  if (!order) return 0
  const weight = order.weight_g ?? order.WeightG ?? 0
  return weight * 10
}

// 已补定金 = 当前订单定金 - 基础定金（小于0按0处理）
const getSupplementDepositDisplay = (order) => {
  if (!order) return 0
  const total = order.deposit ?? order.Deposit ?? 0
  const base = getBaseDeposit(order)
  const extra = total - base
  return extra > 0 ? extra : 0
}

// 列表中“金额/定金”展示专用：
// - 交易记录: 使用基础定金
// - 其他记录: 使用原始金额
const getRecordAmountDisplay = (record) => {
  if (!record) return 0
  if (record.type === 'trade' && record.order) {
    return getBaseDeposit(record.order)
  }
  return record.amount || 0
}

const calcOrderMarginRate = (order) => {
  if (!order) return null
  const base = getBaseDeposit(order)
  if (!base) return null

  const extra = getSupplementDepositDisplay(order)

  // 盈利：
  // - 持仓：按当前价计算浮动盈亏
  // - 已结算/已平仓：优先使用 settled_pnl / SettledPnL，其次回退到 pnl_float / PnLFloat
  const rawStatus = order.status || order.Status || ''
  const status = String(rawStatus).toLowerCase()
  let pnl = 0
  if (status === 'holding') {
    pnl = calcOrderPnl(order)
  } else if (status === 'settled' || status === 'closed') {
    if (typeof order.settled_pnl === 'number') pnl = order.settled_pnl
    else if (typeof order.SettledPnL === 'number') pnl = order.SettledPnL
    else if (typeof order.pnl_float === 'number') pnl = order.pnl_float
    else if (typeof order.PnLFloat === 'number') pnl = order.PnLFloat
  } else {
    pnl = calcOrderPnl(order)
  }

  // 定金率 = (定金 + 已补定金 + 盈利) / 定金 × 100%
  return ((base + extra + pnl) / base) * 100
}

const holdingMargin = computed(() => {
  if (!holdingOrders.value || holdingOrders.value.length === 0) return 0
  return holdingOrders.value.reduce((sum, order) => {
    const deposit = order.deposit || 0
    const pnl = calcOrderPnl(order)
    return sum + deposit + pnl
  }, 0)
})

const totalDeposit = computed(() => {
  const available = userInfo.value.available_deposit || 0
  const pendingRefund = pendingRefundDeposit.value || 0
  return available + holdingMargin.value + pendingRefund
})

// 目标定金率（百分比），默认100，可通过系统配置 auto_supplement_target 覆盖
const targetMarginRate = ref(100)

const tradeType = ref('long_buy')
const tradeStatus = ref('holding')

const tradeRecords = computed(() => {
  if (activeTab.value !== 'trade') return []
  return records.value.filter((r) => {
    const t = (r.type || '').toLowerCase()
    if (t !== 'supplement_deposit' && t !== 'supplement') return false

    // 按锁价买料 / 锁价卖料过滤（如资金流水中包含订单方向）
    const otRaw = (r.order_type || '').toLowerCase()
    if (!otRaw || !tradeType.value) return true

    // 兼容 buy/sell 与 long_buy/short_sell
    const ot =
      otRaw === 'buy' ? 'long_buy' : otRaw === 'sell' ? 'short_sell' : otRaw

    return ot === tradeType.value
  })
})

const visibleRecords = computed(() => {
  if (activeTab.value === 'trade') {
    return tradeRecords.value
  }
  return records.value
})

const getTradeRecordPnl = (record) => {
  if (!record || !record.order) return 0
  const order = record.order
  const rawStatus = order.status || order.Status || ''
  const status = String(rawStatus).toLowerCase()
  if (status === 'holding') {
    return calcOrderPnl(order)
  }
  if (status === 'settled' || status === 'closed') {
    if (typeof order.settled_pnl === 'number') return order.settled_pnl
    if (typeof order.SettledPnL === 'number') return order.SettledPnL
    if (typeof order.pnl_float === 'number') return order.pnl_float
    if (typeof order.PnLFloat === 'number') return order.PnLFloat
  }
  return 0
}

const getOrderDisplayPnl = (order) => {
  if (!order) return 0
  const rawStatus = order.status || order.Status || ''
  const status = String(rawStatus).toLowerCase()
  if (status === 'holding') {
    return calcOrderPnl(order)
  }
  if (status === 'settled' || status === 'closed') {
    if (typeof order.settled_pnl === 'number') return order.settled_pnl
    if (typeof order.SettledPnL === 'number') return order.SettledPnL
    if (typeof order.pnl_float === 'number') return order.pnl_float
    if (typeof order.PnLFloat === 'number') return order.PnLFloat
  }
  return 0
}

const tradeTotalWeight = computed(() => {
  if (activeTab.value !== 'trade') return 0
  return tradeRecords.value.reduce((sum, record) => {
    const order = record.order || {}
    const weight = order.weight_g ?? order.WeightG ?? 0
    return sum + weight
  }, 0)
})

const tradeAvgLockedPrice = computed(() => {
  if (activeTab.value !== 'trade') return 0
  const list = tradeRecords.value
  if (!list.length) return 0
  const totalWeight = list.reduce((sum, record) => {
    const order = record.order || {}
    const weight = order.weight_g ?? order.WeightG ?? 0
    return sum + weight
  }, 0)
  if (!totalWeight) return 0
  const totalLocked = list.reduce((sum, record) => {
    const order = record.order || {}
    const locked = order.locked_price ?? order.LockedPrice ?? 0
    const weight = order.weight_g ?? order.WeightG ?? 0
    return sum + locked * weight
  }, 0)
  return totalLocked / totalWeight
})

const tradeTotalPnlDisplay = computed(() => {
  if (activeTab.value !== 'trade') return 0
  return tradeRecords.value.reduce((sum, record) => sum + getTradeRecordPnl(record), 0)
})

const calcOrderNeedSupplement = (order) => {
  if (!order) return 0
  const base = getBaseDeposit(order)
  if (!base) return 0

  const rawStatus = order.status || order.Status || ''
  const status = String(rawStatus).toLowerCase()
  if (status !== 'holding') return 0

  const rate = calcOrderMarginRate(order) ?? 0
  const targetRate = targetMarginRate.value || 100
  if (rate >= targetRate) return 0

  // ΔA = 定金 / 100 × (R_target - R)
  const delta = (base * (targetRate - rate)) / 100
  return delta > 0 ? delta : 0
}

const tradeTotalNeedSupplement = computed(() => {
  if (activeTab.value !== 'trade') return 0
  return tradeRecords.value.reduce((sum, record) => {
    const order = record.order || {}
    return sum + calcOrderNeedSupplement(order)
  }, 0)
})

const isOrderHolding = (order) => {
  if (!order) return false
  const rawStatus = order.status || order.Status || ''
  const status = String(rawStatus).toLowerCase()
  return status === 'holding'
}

const getOrderTypeText = (type) => {
  if (type === 'long_buy') return '锁价买料'
  if (type === 'short_sell') return '锁价卖料'
  return type || ''
}

const loadHoldingOrders = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.ORDERS, { params: { status: 'holding' } })
    const list = data.orders || data.list || []
    holdingOrders.value = list.map((o) => ({
      order_id: o.order_id || o.OrderID || o.id || '',
      type: o.type || o.Type || '',
      locked_price: o.locked_price ?? o.LockedPrice ?? 0,
      weight_g: o.weight_g ?? o.WeightG ?? 0,
      deposit: o.deposit ?? o.Deposit ?? 0,
      pnl_float: o.pnl_float ?? o.PnLFloat ?? 0,
      status: o.status || o.Status || '',
      created_at: o.created_at || o.CreatedAt || null
    }))
  } catch (error) {
    console.error('加载持仓订单失败:', error)
  }
}

const loadPendingRefundDeposit = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.WITHDRAWS, {
      params: {
        limit: 100,
        offset: 0
      }
    })
    const list = data.withdraws || data.list || []
    const sum = list.reduce((total, w) => {
      const status = String(w.status || w.Status || '').toLowerCase()
      if (status === 'approved') {
        const amount = w.amount ?? w.Amount ?? 0
        return total + (amount || 0)
      }
      return total
    }, 0)
    pendingRefundDeposit.value = sum
  } catch (error) {
    console.error('加载待退定金失败:', error)
  }
}

// 加载目标定金率(%)，来自系统配置 auto_supplement_target
const loadTargetMarginRate = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.CONFIG)
    const list = data.configs || data.list || []
    const item = list.find((c) => {
      const key = c.key || c.Key || ''
      return key === 'auto_supplement_target'
    })
    if (item) {
      const raw = item.value || item.Value || ''
      const num = parseFloat(raw)
      if (!Number.isNaN(num) && num > 0) {
        targetMarginRate.value = num
      }
    }
  } catch (error) {
    console.error('加载目标定金率失败:', error)
  }
}

const showDeposit = ref(false)
const showWithdraw = ref(false)
const showBankCardPicker = ref(false)
const currentPickerType = ref('deposit')
const selectedBankCardText = ref('')
const bankCards = ref([])

const depositForm = ref({
  amount: '',
  note: ''
})

const withdrawForm = ref({
  amount: '',
  bank_card_id: '',
  note: ''
})

const paymentInfo = ref({
  bank_card: null,
  wechat_qr: '',
  alipay_qr: ''
})

// 快捷金额
const quickAmounts = ref([5000, 6000, 10000, 15000, 20000, 50000, 100000, 200000])

// 付款账户选择
const showPaymentCardPicker = ref(false)
const selectedPaymentCard = ref(null)
const paymentCardColumns = ref([])

// 支付凭证
const voucherFiles = ref([])
const voucherUrl = ref('')

// 协议
const agreeProtocol = ref(false)

// 详情弹窗
const showDepositDetailDialog = ref(false)
const currentDetailRecord = ref(null)

// 锁价订单详情（查看料单）
const showTradeDetailDialog = ref(false)
const currentDetailOrder = ref(null)

// 获取记录类型文本
const getRecordTypeText = (type) => {
  const types = {
    deposit: '付定金',
    withdraw: '退定金',
    trade: '补定金',
    buy: '买入',
    sell: '卖出',
    profit: '盈利',
    loss: '亏损',
    commission: '提成',
    supplement_deposit: '补定金',
    supplement: '补定金'
  }
  return types[type] || type
}

// 获取记录类型样式类
const getRecordTypeClass = (type) => {
  const classMap = {
    deposit: 'type-deposit',
    withdraw: 'type-withdraw',
    buy: 'type-buy',
    sell: 'type-sell',
    profit: 'type-profit',
    loss: 'type-loss'
  }
  return classMap[type] || 'type-default'
}

// 获取状态文本
const getStatusText = (status) => {
  if (!status) return ''
  const normalized = String(status).toLowerCase()
  const statusMap = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝',
    paid: '已打款',
    success: '成功',
    failed: '失败',
    holding: '待结算',
    settled: '已完结',
    closed: '已完结'
  }
  return statusMap[normalized] || status
}

// 格式化快捷金额
const formatQuickAmount = (amount) => {
  if (amount >= 10000) {
    return (amount / 10000) + '万'
  }
  return amount
}

// 选择付款账户
const onSelectPaymentCard = (value) => {
  const card = bankCards.value.find(c => (c.id || c.ID) === value.value)
  if (card) {
    selectedPaymentCard.value = {
      id: card.id || card.ID,
      card_holder: card.card_holder || card.CardHolder,
      bank_name: card.bank_name || card.BankName,
      card_number: card.card_number || card.CardNumber
    }
  }
  showPaymentCardPicker.value = false
}

// 上传凭证图片
const afterReadVoucher = async (file) => {
  try {
    showToast('正在处理图片...')
    
    // 处理单个或多个文件
    const files = Array.isArray(file) ? file : [file]
    
    // 初始化数组
    if (!voucherUrl.value) {
      voucherUrl.value = []
    }
    if (typeof voucherUrl.value === 'string') {
      voucherUrl.value = [voucherUrl.value]
    }
    
    for (const f of files) {
      const compressed = await compressImage(f.file)
      
      // 检查单张图片大小
      const sizeKB = Math.round(compressed.length / 1024)
      if (sizeKB > 800) {
        showToast(`图片过大(${sizeKB}KB)，请重新选择`)
        continue
      }
      
      voucherUrl.value.push(compressed)
    }
    
    // 检查总大小
    const totalSize = voucherUrl.value.join(',').length
    const totalSizeKB = Math.round(totalSize / 1024)
    console.log('凭证总大小:', totalSizeKB, 'KB')
    
    if (totalSizeKB > 2000) {
      showToast('凭证图片总大小超限，请减少数量或降低分辨率')
      return
    }
    
    showToast('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    showToast('图片处理失败')
  }
}

// 查看记录详情
const showRecordDetail = (record) => {
  // 显示详情对话框
  showDepositDetailDialog.value = true
  currentDetailRecord.value = record
}

// 查看料单：从交易记录或补定金流水中单独查看锁价订单详情
const openTradeDetail = async (source) => {
  if (!source) return

  // 直接传入订单对象的情况（交易记录）
  if (source.locked_price !== undefined && source.weight_g !== undefined) {
    currentDetailOrder.value = source
    showTradeDetailDialog.value = true
    return
  }

  // 补定金流水：通过 order_id 拉取订单详情
  const orderId = source.order_id || source.OrderID || source.id
  if (!orderId) return

  try {
    const detail = await request.get(API_ENDPOINTS.ORDER_DETAIL.replace(':id', orderId))
    const raw = detail.order || detail || {}
    currentDetailOrder.value = {
      order_id: raw.order_id || raw.OrderID || raw.id || '',
      type: raw.type || raw.Type || '',
      status: raw.status || raw.Status || '',
      locked_price: raw.locked_price ?? raw.LockedPrice ?? 0,
      weight_g: raw.weight_g ?? raw.WeightG ?? 0,
      deposit: raw.deposit ?? raw.Deposit ?? 0,
      pnl_float: raw.pnl_float ?? raw.PnLFloat ?? 0,
      settled_pnl: raw.settled_pnl ?? raw.SettledPnL ?? 0,
      created_at: raw.created_at || raw.CreatedAt || null
    }
    showTradeDetailDialog.value = true
  } catch (error) {
    console.error('加载订单详情失败:', error)
    showToast('加载料单详情失败')
  }
}

// 将后端存储的图片字段解析为 URL 列表
const parseImageUrls = (raw) => {
  if (!raw) return []
  const str = String(raw).trim()
  if (!str) return []

  // 优先匹配一个字段里包含的多个 data:image...base64,... 段
  const dataUrlMatches = str.match(/data:image[^,]*,[^,]+/g)
  if (dataUrlMatches && dataUrlMatches.length > 0) {
    return dataUrlMatches.map((s) => s.trim())
  }

  // 非 Data URL 情况，再按逗号拆分多张图片
  if (str.includes(',')) {
    return str
      .split(',')
      .map((s) => s.trim())
      .filter((s) => s.length > 0)
  }

  // 单一 URL 或裸 base64
  return [str]
}

// 规范图片 URL：支持 http(s)、/path、本地 base64
const normalizeImageUrl = (raw) => {
  if (!raw) return ''
  const url = String(raw).trim()
  if (!url) return ''

  if (
    url.startsWith('http://') ||
    url.startsWith('https://') ||
    url.startsWith('data:') ||
    url.startsWith('/')
  ) {
    return url
  }

  // 兜底：看起来像裸的 base64 内容，补上 jpeg 前缀
  return `data:image/jpeg;base64,${url}`
}

// 获取凭证URL数组
const getVoucherUrls = (voucherUrl) => {
  const urls = parseImageUrls(voucherUrl)
    .map((url) => normalizeImageUrl(url))
    .filter((url) => url && url.length > 0)
  return urls
}

// 预览凭证
const previewVoucher = (voucherUrl, startPosition = 0) => {
  const urls = getVoucherUrls(voucherUrl)
  showImagePreview({
    images: urls,
    startPosition: startPosition
  })
}

// 压缩图片
const compressImage = (file, maxWidth = 600, quality = 0.6) => {
  return new Promise((resolve) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        let width = img.width
        let height = img.height
        
        // 更激进的压缩策略
        if (width > maxWidth) {
          height = (height * maxWidth) / width
          width = maxWidth
        }
        
        canvas.width = width
        canvas.height = height
        
        const ctx = canvas.getContext('2d')
        ctx.drawImage(img, 0, 0, width, height)
        
        // 使用更低的质量
        let compressedBase64 = canvas.toDataURL('image/jpeg', quality)
        
        // 如果还是太大，进一步降低质量
        if (compressedBase64.length > 500000) { // 500KB
          compressedBase64 = canvas.toDataURL('image/jpeg', 0.4)
        }
        
        console.log('压缩后图片大小:', Math.round(compressedBase64.length / 1024), 'KB')
        resolve(compressedBase64)
      }
      img.src = e.target.result
    }
    reader.readAsDataURL(file)
  })
}

// 加载用户信息
const loadUserInfo = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.USER_PROFILE)
    userInfo.value = {
      available_deposit: data.available_deposit || 0,
      used_deposit: data.used_deposit || 0
    }
  } catch (error) {
    console.error('加载用户信息失败:', error)
  }
}

// 加载资金流水
const loadRecords = async () => {
  try {
    let list = []
    
    if (activeTab.value === 'deposit') {
      // 加载充值记录
      const data = await request.get(API_ENDPOINTS.DEPOSITS)
      const deposits = data.deposits || []
      
      // 转换为统一格式
      list = deposits.map(d => ({
        id: d.ID || d.id,
        type: 'deposit',
        amount: d.Amount || d.amount,
        status: d.Status || d.status,
        created_at: d.CreatedAt || d.created_at,
        voucher_url: d.VoucherURL || d.voucher_url,
        method: d.Method || d.method,
        review_note: d.ReviewNote || d.review_note,
        reviewed_at: d.ReviewedAt || d.reviewed_at,
        description: d.ReviewNote || d.review_note || ''
      }))
    } else if (activeTab.value === 'withdraw') {
      // 加载提现记录
      const data = await request.get(API_ENDPOINTS.WITHDRAWS)
      const withdraws = data.withdraws || []
      
      list = withdraws.map(w => ({
        id: w.ID || w.id,
        type: 'withdraw',
        amount: -(w.Amount || w.amount),
        status: w.Status || w.status,
        created_at: w.CreatedAt || w.created_at,
        bank_card_id: w.BankCardID || w.bank_card_id,
        review_note: w.ReviewNote || w.review_note,
        reviewed_at: w.ReviewedAt || w.reviewed_at,
        paid_at: w.PaidAt || w.paid_at,
        voucher_url: w.VoucherURL || w.voucher_url,
        description: w.UserNote || w.user_note || w.ReviewNote || w.review_note || ''
      }))
    } else {
      // 加载所有资金流水
      const data = await request.get(API_ENDPOINTS.FUND_FLOW)
      const logs = data.logs || []

      const mapped = logs.map((log) => ({
        id: log.ID || log.id,
        type: log.Type || log.type,
        amount: log.Amount || log.amount,
        before_balance: log.AvailableBefore || log.available_before,
        after_balance: log.AvailableAfter || log.available_after,
        created_at: log.CreatedAt || log.created_at,
        description: log.Note || log.note || '',
        // 可能存在的订单关联字段
        order_id: log.OrderID || log.order_id || log.OrderId || '',
        order_type: log.OrderType || log.order_type || log.OrderSide || log.order_side || ''
      }))

      if (activeTab.value === 'trade') {
        // 补定金 Tab：仅展示补定金类流水
        list = mapped.filter((log) => {
          const t = (log.type || '').toLowerCase()
          return t === 'supplement_deposit' || t === 'supplement'
        })
      } else {
        list = mapped
      }
    }
    
    console.log('加载的记录:', list)
    records.value = list

    // 在补定金 Tab 中，为补定金流水按需补充订单方向（order_type），用于锁价买料/卖料筛选
    if (activeTab.value === 'trade') {
      const tasks = records.value
        .filter((r) => {
          const t = (r.type || '').toLowerCase()
          return (
            (t === 'supplement_deposit' || t === 'supplement') &&
            !r.order_type &&
            r.order_id
          )
        })
        .map(async (r) => {
          try {
            const detail = await request.get(
              API_ENDPOINTS.ORDER_DETAIL.replace(':id', r.order_id)
            )
            const raw = detail.order || detail || {}
            r.order_type = raw.type || raw.Type || ''
          } catch (error) {
            console.error('补定金流水加载订单方向失败:', error)
          }
        })

      if (tasks.length > 0) {
        try {
          await Promise.all(tasks)
        } catch (e) {
          // 单条失败不影响整体
        }
      }
    }

    finished.value = true
  } catch (error) {
    console.error('加载资金流水失败:', error)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

// 加载银行卡列表
const loadBankCards = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.BANK_CARDS)
    console.log('银行卡数据:', data)
    bankCards.value = data.cards || data.list || []
    console.log('解析后的银行卡列表:', bankCards.value)
    
    // 初始化付款账户选择器
    paymentCardColumns.value = bankCards.value.map(card => ({
      text: `${card.bank_name || card.BankName} (*${(card.card_number || card.CardNumber || '').slice(-4)})`,
      value: card.id || card.ID
    }))
    
    // 默认选择第一张卡
    if (bankCards.value.length > 0) {
      const firstCard = bankCards.value[0]
      selectedPaymentCard.value = {
        id: firstCard.id || firstCard.ID,
        card_holder: firstCard.card_holder || firstCard.CardHolder,
        bank_name: firstCard.bank_name || firstCard.BankName,
        card_number: firstCard.card_number || firstCard.CardNumber
      }
    }
  } catch (error) {
    console.error('加载银行卡失败:', error)
  }
}

// 打开银行卡选择器
const openBankCardPicker = (type) => {
  currentPickerType.value = type
  showBankCardPicker.value = true
}

// 选择银行卡
const selectBankCard = (card) => {
  const cardId = card.id || card.ID
  const bankName = card.bank_name || card.BankName
  const cardNumber = card.card_number || card.CardNumber || ''
  
  // 只用于提现
  withdrawForm.value.bank_card_id = cardId
  selectedBankCardText.value = `${bankName} (*${cardNumber.slice(-4)})`
  showBankCardPicker.value = false
}

// 跳转到添加银行卡
const goToAddCard = () => {
  showBankCardPicker.value = false
  showDeposit.value = false
  showWithdraw.value = false
  window.location.href = '#/bank-cards'
}

// 付定金
const onDeposit = async () => {
  try {
    // 验证
    if (!depositForm.value.amount) {
      showToast('请输入金额')
      return
    }
    
    if (!selectedPaymentCard.value) {
      showToast('请选择付款账户')
      return
    }
    
    if (!agreeProtocol.value) {
      showToast('请阅读并同意协议')
      return
    }
    
    // 处理凭证URL（支持多张图片）
    let voucherUrlString = ''
    if (voucherUrl.value) {
      if (Array.isArray(voucherUrl.value)) {
        // 多张图片用逗号分隔
        voucherUrlString = voucherUrl.value.join(',')
      } else {
        voucherUrlString = voucherUrl.value
      }
    }
    
    const requestData = {
      amount: parseFloat(depositForm.value.amount),
      method: 'bank',
      voucher_url: voucherUrlString,
      note: depositForm.value.note || ''
    }
    
    console.log('付定金请求数据:', requestData)
    
    await request.post(API_ENDPOINTS.DEPOSIT_CREATE, requestData)
    
    showToast('付定金申请已提交，等待审核')
    
    // 重置表单
    showDeposit.value = false
    depositForm.value = { amount: '', note: '' }
    voucherFiles.value = []
    voucherUrl.value = []
    agreeProtocol.value = false
    
    loadUserInfo()
    onRefresh()
  } catch (error) {
    console.error('付定金失败:', error)
    console.error('错误详情:', error.response?.data)
    const errorMsg = error.response?.data?.error || error.response?.data?.message || '付定金失败'
    showToast(errorMsg)
  }
}

// 提现
const onWithdraw = async () => {
  try {
    await request.post(API_ENDPOINTS.WITHDRAW_CREATE, {
      amount: parseFloat(withdrawForm.value.amount),
      bank_card_id: withdrawForm.value.bank_card_id,
      note: withdrawForm.value.note || ''
    })
    
    showToast('提现申请已提交，等待审核')
    showWithdraw.value = false
    withdrawForm.value = { amount: '', bank_card_id: '', note: '' }
    selectedBankCardText.value = ''
    
    loadUserInfo()
    onRefresh()
  } catch (error) {
    console.error('提现失败:', error)
  }
}

// 下拉刷新
const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadRecords()
  loadUserInfo()
  loadPendingRefundDeposit()
}

const onClickSupplement = async (order) => {
  if (!order) return
  const amount = calcOrderNeedSupplement(order)
  if (!amount) {
    showToast('当前订单无需补定金')
    return
  }

  try {
    await showDialog({
      title: '确认补定金',
      message: `该订单需补定金：¥${formatMoney(amount)}，是否立即从可用定金中补充？`,
      showCancelButton: true
    })
  } catch (error) {
    // 取消
    return
  }

  try {
    loading.value = true
    await request.post(API_ENDPOINTS.SUPPLEMENTS, {
      order_id: order.db_id || order.ID || order.id,
      amount: amount
    })
    showToast('补定金成功')
    loadUserInfo()
    onRefresh()
  } catch (error) {
    console.error('补定金失败:', error)
    const msg = error.response?.data?.error || error.response?.data?.message || '补定金失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}

const showSettleDialog = ref(false)
const settleOrder = ref(null)
const settlePayPassword = ref('')

const openSettleDialog = (order) => {
  if (!order) return
  settleOrder.value = order
  settlePayPassword.value = ''
  showSettleDialog.value = true
}

const confirmSettle = async () => {
  if (!settleOrder.value) return
  const order = settleOrder.value
  const price = calcOrderCurrentPrice(order)
  if (!price) {
    showToast('当前价格不可用，暂时无法结算')
    return
  }

  try {
    loading.value = true
    await request.post(
      API_ENDPOINTS.ORDER_SETTLE.replace(':id', order.order_id || order.OrderID || ''),
      {
        settle_price: price,
        pay_password: settlePayPassword.value
      }
    )
    showToast('结算成功')
    showSettleDialog.value = false
    settleOrder.value = null
    loadUserInfo()
    onRefresh()
  } catch (error) {
    console.error('结算失败:', error)
    const msg = error.response?.data?.error || error.response?.data?.message || '结算失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}

// 加载收款信息
const loadPaymentInfo = () => {
  try {
    const settings = localStorage.getItem('payment_settings')
    if (settings) {
      paymentInfo.value = JSON.parse(settings)
    }
  } catch (error) {
    console.error('加载收款信息失败:', error)
  }
}

onMounted(() => {
  quoteStore.connectWebSocket()
  loadUserInfo()
  loadHoldingOrders()
  loadRecords()
  loadBankCards()
  loadPaymentInfo()
  loadPendingRefundDeposit()
  loadTargetMarginRate()
})
</script>

<style scoped>
.funds-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.balance-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 30px 20px;
  color: #fff;
  margin-bottom: 10px;
}

.balance-item {
  margin-bottom: 20px;
}

.balance-item .label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.balance-item .amount {
  font-size: 32px;
  font-weight: bold;
}

.balance-row {
  display: flex;
  justify-content: space-between;
}

.balance-row .balance-item .amount {
  font-size: 20px;
}

.actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.actions .van-button {
  flex: 1;
}

/* 卡片式记录列表 */
.record-card {
  background: #fff;
  margin: 12px;
  padding: 16px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: all 0.3s;
}

.record-card:active {
  transform: scale(0.98);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.record-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.record-type-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 500;
}

.type-deposit {
  background: #fff3e0;
  color: #ff9800;
}

.type-withdraw {
  background: #e3f2fd;
  color: #2196f3;
}

.type-buy {
  background: #f3e5f5;
  color: #9c27b0;
}

.type-sell {
  background: #e8f5e9;
  color: #4caf50;
}

.type-profit {
  background: #ffebee;
  color: #f44336;
}

.type-loss {
  background: #fce4ec;
  color: #e91e63;
}

.type-default {
  background: #f5f5f5;
  color: #666;
}

.type-trade-long {
  background: #f56c6c;
  color: #ffffff;
}

.type-trade-short {
  background: #67c23a;
  color: #ffffff;
}

.record-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.record-status.pending {
  background: #fff3e0;
  color: #ff9800;
}

.record-status.approved,
.record-status.success,
.record-status.paid {
  background: #e8f5e9;
  color: #4caf50;
}

.record-status.rejected,
.record-status.failed {
  background: #ffebee;
  color: #f44336;
}

.record-card-body {
  margin-bottom: 12px;
}

.record-info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 14px;
}

.record-info-row .label {
  color: #999;
}

.record-info-row .value {
  color: #333;
  font-weight: 500;
  text-align: right;
  max-width: 60%;
}

.record-info-row .amount-text {
  font-size: 16px;
  font-weight: bold;
}

.record-info-row .amount-text.income {
  color: #f56c6c;
}

.record-info-row .amount-text.expense {
  color: #67c23a;
}

.record-info-row .desc-text {
  font-size: 13px;
  color: #666;
}

.record-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.record-actions {
  display: inline-flex;
  gap: 8px;
}

.view-detail {
  font-size: 13px;
  color: #666;
  text-align: right;
  color: #999;
  font-size: 12px;
}

.popup-content {
  padding: 20px;
}

.trade-filters {
  padding: 8px 12px 0;
}

.trade-type-tabs,
.trade-status-tabs {
  display: flex;
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
  margin-bottom: 8px;
}

.trade-type-tab,
.trade-status-tab {
  flex: 1;
  text-align: center;
  padding: 6px 0;
  font-size: 14px;
  color: #666;
}

.trade-type-tab.active,
.trade-status-tab.active {
  background: #1989fa;
  color: #fff;
  font-weight: 500;
}

/* 补定金 Tab：锁价买料 / 锁价卖料 激活色 */
.trade-type-long.active {
  background: #f56c6c;
  color: #fff;
}

.trade-type-short.active {
  background: #67c23a;
  color: #fff;
}

.trade-summary-card {
  margin: 0 12px 8px;
  padding: 10px 12px;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  justify-content: space-between;
}

.trade-summary-item {
  flex: 1;
  text-align: center;
}

.trade-summary-item .label {
  font-size: 12px;
  opacity: 0.85;
  margin-bottom: 4px;
}

.trade-summary-item .value {
  font-size: 15px;
  font-weight: 600;
}

.trade-summary-item .value.profit {
  color: #f56c6c;
}

.trade-summary-item .value.loss {
  color: #67c23a;
}

.record-actions {
  display: inline-flex;
  gap: 8px;
  margin-right: 8px;
}

.detail-actions {
  margin: 16px 16px 0;
}

.pnl-profit-cell .van-cell__value {
  color: #f56c6c;
}

.pnl-loss-cell .van-cell__value {
  color: #67c23a;
}

.record-item {
  background: #fff;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.record-type {
  font-size: 16px;
  font-weight: 500;
}

.record-amount {
  font-size: 18px;
  font-weight: bold;
}

.record-amount.income {
  color: #f56c6c;
}

.record-amount.expense {
  color: #67c23a;
}

.record-body {
  font-size: 14px;
  color: #999;
}

.record-desc {
  margin-bottom: 4px;
}

.popup-content {
  padding: 20px;
}

.popup-header {
  text-align: center;
  margin-bottom: 20px;
}

.popup-header h3 {
  margin: 0;
  font-size: 18px;
}

.bank-card-list {
  padding: 16px;
}

.bank-card-item {
  padding: 16px;
  background: #f7f8fa;
  border-radius: 8px;
  margin-bottom: 12px;
}

.bank-name {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
}

.card-number {
  font-size: 14px;
  color: #666;
}

.empty-tip {
  text-align: center;
  color: #999;
  padding: 40px 0;
}

.empty {
  padding: 40px 0;
}

/* 付定金弹窗样式 */
.deposit-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.deposit-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  padding-bottom: 80px;
}

.amount-section {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.amount-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.amount-input {
  font-size: 24px;
  font-weight: bold;
  padding: 0;
}

.amount-input :deep(.van-field__control) {
  font-size: 24px;
  font-weight: bold;
}

.quick-amounts {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.quick-amounts .van-button {
  border-radius: 4px;
}

.section {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  color: #ee0a24;
  margin-bottom: 12px;
}

.section-title.required::after {
  content: '';
}

.tip-text {
  font-size: 12px;
  color: #ee0a24;
  margin-bottom: 12px;
}

.card-info {
  background: #f7f8fa;
  border-radius: 8px;
  padding: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.info-row:last-child {
  border-bottom: none;
}

.info-row .label {
  color: #666;
  font-size: 14px;
}

.info-row .value {
  color: #333;
  font-size: 14px;
  font-weight: 500;
}

.tip-content {
  font-size: 12px;
  color: #999;
  line-height: 1.6;
}

.agreement {
  margin: 16px 0;
}

.submit-btn {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16px;
  background: #fff;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05);
}

/* 详情弹窗样式 */
.detail-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.income-cell :deep(.van-cell__value) {
  color: #f56c6c !important;
  font-weight: bold;
}

.expense-cell :deep(.van-cell__value) {
  color: #67c23a !important;
  font-weight: bold;
}

.voucher-section {
  margin-top: 16px;
  background: #fff;
  border-radius: 8px;
  padding: 16px;
}

.voucher-section .section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 12px;
  color: #333;
}

.voucher-images {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.voucher-images .van-image {
  border-radius: 4px;
  cursor: pointer;
}
</style>
