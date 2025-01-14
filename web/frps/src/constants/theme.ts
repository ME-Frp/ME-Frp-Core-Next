export const themeColors = {
  primary: '#2196F3',
  primaryHover: '#42A5F5',
  primaryPressed: '#1976D2',
  primarySuppl: '#90CAF9',
} as const

export const themeOverrides = {
  common: {
    primaryColor: '#2196F3',
    primaryColorHover: '#42A5F5',
    primaryColorPressed: '#1976D2',
    primaryColorSuppl: '#90CAF9'
  },
  Button: {
    textColorPrimary: '#FFFFFF',
    textColorHoverPrimary: '#FFFFFF',
    textColorPressedPrimary: '#FFFFFF',
    textColorFocusPrimary: '#FFFFFF',
  },
  Menu: {
    itemTextColorActive: themeColors.primary,
    itemIconColorActive: themeColors.primary,
    itemTextColorActiveHover: themeColors.primaryHover,
    itemIconColorActiveHover: themeColors.primaryHover,
  }
} as const 