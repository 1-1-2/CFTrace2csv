import pandas as pd


def write2xlsx(df, filename='cf_trace.xlsx'):
    with pd.ExcelWriter(filename, engine='openpyxl') as writer:
        df.to_excel(writer, sheet_name='cf_trace')

        # 自动计算并调整列宽
        max_width = 80
        for sheets in writer.sheets.values():
            for column in sheets.columns:
                max_length = 0
                column_letter = column[0].column_letter  # 获取列的字母标识
                for cell in column:
                    try:
                        this_length = len(str(cell.value))
                        if this_length > max_length:
                            max_length = this_length
                    except Exception as e:
                        print(e)
                # 添加一些额外的宽度，同时限制最大宽度，第一列通常加粗，再宽点
                if column_letter == 'A':
                    adjusted_width = min(max_length * 1.2 + 2, max_width)
                else:
                    adjusted_width = min(max_length + 2, max_width)
                sheets.column_dimensions[column_letter].width = adjusted_width


def line2dict(line):
    data = {}
    for field in line.split(','):
        if '=' in field:
            try:
                key, value = field.split('=')
                data[key] = value
            except Exception as e:
                print(e, line[:100] + "..." if len(line) > 100 else line)
                return {'CDN_IP': data['CDN_IP']}

    return data


if __name__ == '__main__':
    with open('cf_trace.csv', 'r') as f:
        table = f.readlines()

    table_parsed = list()
    for line in table:
        table_parsed.append(line2dict(line))

    df = pd.DataFrame(table_parsed)

    # 有的节点会记录本地 ip，应为 ip 列的众数
    loc_ip = df.ip.mode()[0]
    # 将 table_parsed.ip 与本地 ip 相同的用空字符串填充
    df.ip[df.ip == loc_ip] = None

    # colo2location
    df_colo = pd.read_csv('ColoList.csv', encoding='GBK')
    # 将 '国家CH' 和 '地区CH' 拼合到 'CH' 列
    df_colo['CH'] = df_colo['国家CH'] + ',' + df_colo['地区CH']
    # 将 df.colo 根据 df_colo 的映射，存放到 df.ts 列
    df['ts'] = df.colo.map(df_colo.set_index('三字码')['CH'])

    # write2xlsx(df)
