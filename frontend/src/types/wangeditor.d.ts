// @wangeditor/editor-for-vue 的 package.json "exports" 未暴露类型入口，
// 这里补充模块声明，让 TS 能解析其类型（运行时类型来自 @wangeditor/editor）。
declare module '@wangeditor/editor-for-vue' {
  import type { DefineComponent } from 'vue'
  import type { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'

  export const Editor: DefineComponent<{
    modelValue?: string
    defaultConfig?: Partial<IEditorConfig>
    mode?: 'default' | 'simple'
    style?: Record<string, string> | string
    onOnCreated?: (editor: IDomEditor) => void
    onOnChange?: (editor: IDomEditor) => void
    onOnDestroyed?: (editor: IDomEditor) => void
    onOnFocus?: (editor: IDomEditor) => void
    onOnBlur?: (editor: IDomEditor) => void
  }>

  export const Toolbar: DefineComponent<{
    editor?: IDomEditor | null
    defaultConfig?: Partial<IToolbarConfig>
    mode?: 'default' | 'simple'
  }>
}
