
---

# 【AI 重构与编码指导方案：Bubble Tea + Lipgloss 组件化架构】

**目的：**
指导 AI 按照统一标准，对 Bubble Tea + Lipgloss 项目进行组件化重构，避免不必要的结构发明、过度抽象或偏离设计方向。

本规范**必须被严格遵守**。任何 AI 在生成代码或修改代码前，都要先验证是否符合本规范。

---

# 0. 总体目标

将现有 Bubble Tea 代码重构为：

1. 拆分为 **组件层 (components)**
2. 引入 **轻量布局 DSL (layouts)**
3. 页面的 UI 逻辑分离为 **pages**
4. 样式集中到 **theme**
5. 主 Model 只负责分发消息与管理子组件

不引入：

* 虚拟 DOM
* 复杂 runtime
* 不属于 Bubble Tea 的“控件系统”
* AI 自己杜撰的新框架

---

# 1. 目录结构规范（必须严格遵守）

AI 必须将所有文件按以下结构组织：

```
ui/
  components/
    xxx_component.go
  layouts/
    layout.go
    row.go
    column.go
    section.go
  pages/
    xxx_page.go
  theme/
    style.go
    color.go

model/
  model.go
  msg.go

main.go
```

## 禁止

* 将 UI 与 model 混在同一文件
* 在组件内写 style 定义
* 在 page 内写布局 DSL
* 写任意未在本规范中定义的目录

---

# 2. Bubble Tea 组件模式规范（必须遵守）

每个组件文件 **必须** 遵循如下模式：

```
type XxxComponent struct {
    // 组件状态
}

func NewXxxComponent(...) XxxComponent {
}

func (c XxxComponent) Update(msg tea.Msg) (XxxComponent, tea.Cmd) {
}

func (c XxxComponent) View() string {
}
```

### 组件原则：

1. **组件只能处理自己的状态**
2. **组件不允许访问 Model 的外部字段**
3. **组件的 View() 只能返回 string**
4. **组件的样式必须从 /ui/theme 引入**
5. **组件不允许自行处理 layout（Row/Column）**

   * 布局由 pages 或主 view 处理
   * 组件必须“只关心内容，不关心布局”

---

# 3. Page（页面）规范（必须遵守）

Page 是由组件组合形成的 UI 单元：

```
type Page interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Page, tea.Cmd)
    View() string
}
```

示例结构：

```
type HomePage struct {
    Tasks    TaskList
    Stats    StatsPanel
}

func NewHomePage() HomePage {
    return HomePage{
        Tasks: NewTaskList(),
        Stats: NewStatsPanel(),
    }
}
```

### Page 必须满足：

* 字段只能是组件或简单状态
* View() 只能使用 `/ui/layouts` 提供的布局方法
* 不允许在页面中写具体样式
* 不允许直接构建 strings.Builder

---

# 4. 布局 DSL 规范（必须遵守）

布局只能通过 `/ui/layouts` 中的函数实现，不允许 AI 自己发明布局 API。

布局 DSL 必须包含：

```
Row(blocks ...string) string
Column(blocks ...string) string
Section(title string, body string) string
```

可选（如需要）：

```
Gap(n int) string
Pad(padding int, content string) string
```

### 布局系统规则：

1. **布局只表示关系，不负责样式**
2. **布局不能包含业务逻辑**
3. **布局不得写入硬编码 lipgloss style**

   * 正确：style 统一在 theme 定义
   * 错误：Row() 中出现 `.Border()`

---

# 5. UI 主题规范（必须遵守）

所有 Style **必须在 `/ui/theme`** 中定义。

例如：

```
var HeaderStyle = lipgloss.NewStyle().
    Foreground(colorPrimary).
    Bold(true).
    Padding(1, 2)
```

AI 不得：

* 在组件中创建 `lipgloss.NewStyle()`
* 在页面中写颜色、padding、border 等
* 在布局 DSL 中写任何 style

所有 style 都必须统一从 theme import。

---

# 6. 主 Model 规范（必须遵守）

主 Model 负责：

* 管理当前 Page
* 将消息分发给当前 Page
* 整体 View 使用 layout DSL 进行组合

典型模式：

```
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    newPage, cmd := m.Page.Update(msg)
    m.Page = newPage
    return m, cmd
}

func (m Model) View() string {
    return Column(
        Header(m),
        m.Page.View(),
        Footer(m),
    )
}
```

禁止：

* 主 Model 中拼接字符串
* 在主 View 中出现任何 lipgloss style
* 在主 Model 中访问组件内部状态

---

# 7. 修改现有项目时 AI 必须遵循的流程（必须遵守）

1. **扫描现有视图代码**
2. 标记可拆分的组件（列表、卡片、状态栏、日志区、按钮行等）
3. 在 `/ui/components` 中为每一个组件创建文件
4. 将视图逻辑从主 Model 转移到组件
5. 将布局逻辑移动到 `/ui/layouts`
6. 将样式提取到 `/ui/theme`
7. 将业务状态与 UI 页面解耦（pages）
8. 主 Model 最终只负责页面切换与 msg 分发
9. 不允许生成超出规范的新 API、结构或文件

---

# 8. AI 必须遵守的风格验证 Checklist

在提交任何代码前，AI 必须检查：

### 文件结构 OK？

* 组件在 components
* 页面在 pages
* 布局在 layouts
* 样式在 theme
* 主 Model 在 model

### 组件是否合格？

* 拆了 View？
* 没写 style？
* 没写布局？
* 有 Update()?
* 有 View()?

### 页面是否合格？

* 只使用布局 DSL？
* 不出现 lipgloss style？
* 不直接 string 操作？

### theme 是否唯一？

* 没有组件偷偷创建 style？

### 主 Model 是否“瘦”？

* 没写 UI？
* 没写 lipgloss？
* 只负责 Page 切换？

**如果任一项失败，AI 必须重新修改，而不是继续生成。**

---

# 9. 强制性禁止条目（AI 不允许做的）

AI 绝不能做以下事情：

* 发明新的组件系统（比如 Tabs、TreeView）
* 引入虚拟 DOM 或 diff
* 混合 model 和 view
* 在 layout 或 component 内写 style
* 在 component 内写布局逻辑
* 在页面中拼字符串
* 在 component 中创建 models
* 改变项目结构（不得创建新的目录）
* 使用反射、接口泛滥、不必要抽象
* 引入非 Bubble Tea 技术（cespare/tui、termui 等）

任何违反以上规则的代码都必须拒绝生成。

---
