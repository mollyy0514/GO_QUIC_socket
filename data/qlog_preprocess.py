import os
import csv
import ast
import json
import statistics
import numpy as np
import pandas as pd
from datetime import datetime, timedelta
from pytictoc import TicToc

##### ---------- USER SETTINGS ---------- #####
# database = "/Volumes/mollyT7/MOXA/"
# database = "/home/wmnlab/Documents/r12921105"
database = "/Users/molly/Desktop/"
dates = ["2024-04-11"]
exp_names = {
    "QUIC-inf": (1, ["#{:02d}".format(i + 1) for i in range(1)]),
    }
device_names = ["sm01"]

device_to_port = {"sm00": [5200, 5201], 
                  "sm01": [5202, 5203],
                  "sm02": [5204, 5205]}
num_devices = len(device_names)

##### ---------- USER SETTINGS ---------- #####


# pait: [sent_qlog_file, received_qlog_file, time, port, device, round]
def FindFilePairs(database, date, exp, device):
    file_pairs = []
    # [sent_qlog_file, received_qlog_file, time, port, device, round]
    ul_files = ["", "", "", "", "", ""]
    dl_files = ["", "", "", "", "", ""]
    exp_round, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'raw')
        for root, dirs, files in os.walk(folder_path):
            qlog_files = [file for file in files if file.endswith(".qlog")]
            ul_files = ["", "", "", "", "", ""]
            dl_files = ["", "", "", "", "", ""]
            for file in qlog_files:
                time = file.split("_")[2]
                numbers = file.split("_")[3]
                role = file.split("_")[4]
                if str(ports[0]) in numbers:
                    ul_files[3] = ports[0]
                    if "client" in role:
                        ul_files[0] = os.path.join(root, file)
                    elif "server" in role:
                        ul_files[1] = os.path.join(root, file)
                        ul_files[2] = time
                if str(ports[1]) in numbers:
                    dl_files[3] = ports[1]
                    if "server" in role:
                        dl_files[0] = os.path.join(root, file)
                        dl_files[2] = time
                    elif "client" in role:
                        dl_files[1] = os.path.join(root, file)
            ul_files[4] = device
            dl_files[4] = device
            ul_files[5] = exp_round[1:]
            dl_files[5] = exp_round[1:]
            file_pairs.append(ul_files)
            file_pairs.append(dl_files)
    # print(file_pairs)
    return file_pairs

# Convert .qlof to JSON entry
def QlogToJsonEntry(file_path):
    with open(file_path, 'r') as file:
        lines = file.readlines()

    # Add commas between lines
    json_str = ",".join(lines)
    # Surround the entire string with square brackets to create a JSON array
    json_str = "[" + json_str + "]"
    # Load the JSON array
    json_entry = json.loads(json_str)
    
    return json_entry
# Export JSON entry to .json
def QlogToJson(json_entry, json_file_path):
    with open(json_file_path, 'w') as json_file:
        json.dump(json_entry, json_file, indent=2)
# Convert JSON entry to .csv
def JsonToCsv(json_entry, csv_file_path):
     # Open CSV file for writing
    with open(csv_file_path, 'w', newline='') as csv_file:
        # Create a CSV writer
        csv_writer = csv.writer(csv_file)

        # Write header row based on the keys of the second JSON object (assuming at least two objects are present)
        if len(json_entry) >= 2:
            header = list(json_entry[1].keys())
            csv_writer.writerow(header)

            # Write data rows starting from the second object
            for entry in json_entry[1:]:
                csv_writer.writerow(entry.values())

# Get the connection start time for client or server
def GetStartTime(json_data):
    # unit: ms
    refTime = json_data[0]["trace"]["common_fields"]["reference_time"]
    return refTime
# Add ['epoch_time'] and ['timestamp'] column in df
def ProcessTime(df, reference_time):
    # Extract the "time" values from the DataFrame
    original_times = (df['time'].astype(float))

    # Calculate "epoch_time" and convert to timestamps
    epoch_times = (reference_time + original_times)
    timestamps = pd.to_datetime(epoch_times, unit='ms').dt.strftime('%Y-%m-%d %H:%M:%S.%f')

    df['epoch_time'] = epoch_times
    df['timestamp'] = timestamps

    return df

def insert(df, idx, new_row):
    df1 = df.iloc[:idx, :]
    df2 = df.iloc[idx:, :]
    df_new = pd.concat([df1, new_row, df2], ignore_index=True)
    return df_new

# Mapping RTT to every sent packet
def update_pk_sent_rows(row):
    # print(packet_number_index_map[0])
    acked_ranges = row['acked_ranges']
    smoothed_rtt = row['smoothed_rtt']
    latest_rtt = row['latest_rtt']
    rtt_variance = row['rtt_variance']
    congestion_window = row['congestion_window']

    for ack_range in acked_ranges:
        print(ack_range[0], ack_range[-1])
        start_packet, end_packet = ack_range[0], ack_range[-1]
        start_index = -1
        tmp = start_packet
        while start_index == -1:
            try:
                start_index = packet_number_index_map[tmp]
                print(start_index, packet_number_index_map[tmp])
            except KeyError:
                if (tmp + 1) <= end_packet:
                    tmp += 1
                else:
                    break
        end_index = -1
        tmp = end_packet
        while end_index == -1:
            try:
                end_index = packet_number_index_map[tmp]
                print(end_index, packet_number_index_map[tmp])
            except KeyError:
                if (tmp - 1) >= start_packet:
                    tmp -= 1
                else:
                    break

        if start_index == -1 | end_index == -1:
            continue

        # Update the corresponding rows in pk_sent_rows
        pk_sent_rows.loc[start_index:end_index, 'smoothed_rtt'] = pk_sent_rows.loc[start_index:end_index, 'smoothed_rtt'].fillna(smoothed_rtt)
        pk_sent_rows.loc[start_index:end_index, 'latest_rtt'] = pk_sent_rows.loc[start_index:end_index, 'latest_rtt'].fillna(latest_rtt)
        pk_sent_rows.loc[start_index:end_index, 'congestion_window'] = pk_sent_rows.loc[start_index:end_index, 'congestion_window'].fillna(congestion_window)
        pk_sent_rows.loc[start_index:end_index, 'rtt_variance'] = pk_sent_rows.loc[start_index:end_index, 'rtt_variance'].fillna(rtt_variance)

def find_ul_file(database, date, exp, device):
    ul_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'data')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "processed_rcv" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[0]) in numbers:
                        ul_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return ul_files
def find_dl_file(database, date, exp, device):
    dl_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'data')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "processed_rcv" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[1]) in numbers:
                        dl_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return dl_files
def find_ul_loss_file(database, date, exp, device):
    ul_loss_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'middle')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "lost_pk" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[0]) in numbers:
                        ul_loss_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return ul_loss_files
def find_dl_loss_file(database, date, exp, device):
    dl_loss_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'middle')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "lost_pk" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[1]) in numbers:
                        dl_loss_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return dl_loss_files

def find_ul_sender_file(database, date, exp, device):
    ul_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'raw')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "client.csv" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[0]) in numbers:
                        ul_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return ul_files
def find_dl_sender_file(database, date, exp, device):
    dl_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'raw')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "server.csv" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[1]) in numbers:
                        dl_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return dl_files
def find_ul_rcv_file(database, date, exp, device):
    ul_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'data')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "processed_rcv" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[0]) in numbers:
                        ul_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return ul_files
def find_dl_rcv_file(database, date, exp, device):
    dl_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'data')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "processed_rcv" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[1]) in numbers:
                        dl_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return dl_files
def find_ul_loss_file(database, date, exp, device):
    ul_loss_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'middle')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "lost_pk" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[0]) in numbers:
                        ul_loss_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return ul_loss_files
def find_dl_loss_file(database, date, exp, device):
    dl_loss_files = []
    exp_rounds, exp_list = exp_names[exp]
    ports = device_to_port.get(device, [])
    for exp_round in exp_list:
        folder_path = os.path.join(database, date, exp, device, exp_round, 'middle')
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                if "lost_pk" in file:
                    # Extract the numbers from the file name
                    numbers = file.split("_")[3]
                    if str(ports[1]) in numbers:
                        dl_loss_files.append(os.path.join(root, file))
                        break  # Exit the inner loop once the port is found
    return dl_loss_files

# Classify lost packet
def get_loss_data(lost_df, received_df):
    # Check if each row in ul_lost_df['packet_number'] is present in ul_received_df['packet_number']
    lost_in_received = lost_df['packet_number'].isin(received_df['packet_number'])

    # Get the rows in ul_lost_df where the packet number is present in ul_received_df
    exec_lat_df = lost_df[lost_in_received]

    exec_reorder_df = exec_lat_df[exec_lat_df['trigger'] == 'reordering_threshold']
    exec_time_df = exec_lat_df[exec_lat_df['trigger'] == 'time_threshold']

    # Get the rows in ul_lost_df where the packet number is not present in ul_received_df
    real_lost_df = lost_df[~lost_in_received]

    # Filter ul_lost_df for rows where 'trigger' is 'reordering threshold'
    lost_reorder_df = real_lost_df[real_lost_df['trigger'] == 'reordering_threshold']
    lost_time_df = real_lost_df[real_lost_df['trigger'] == 'time_threshold']

    return exec_lat_df, exec_reorder_df, exec_time_df, real_lost_df, lost_reorder_df, lost_time_df
# Calculate lost packet statistics
def calculate_statistics(lost_reorder_df, lost_time_df, real_lost_df, exec_reorder_df, exec_time_df, exec_lat_df, lost_df, data_df, sent_df):
    statistics_data = [{
        'total_packets': len(sent_df),
        'total_data_packets': len(data_df),
        'original_pkl': len(lost_df),
        'reordering_threshold': len(lost_reorder_df),
        'time_threshold': len(lost_time_df),
        'real_pkl': len(real_lost_df),
        'exec_reordering': len(exec_reorder_df),
        'exec_time': len(exec_time_df),
        'exec_lat': len(exec_lat_df),
        'reordering_pkl_rate(%)': 0 if len(real_lost_df) == 0 else len(lost_reorder_df)*100 / len(real_lost_df),
        'time_pkl_rate(%)': 0 if len(real_lost_df) == 0 else len(lost_time_df)*100 / len(real_lost_df),
        'real_pkl_rate(%)': 0 if len(lost_df) == 0 else len(real_lost_df)*100 / len(lost_df),
        'original_packet_loss_rate(%)': len(lost_df)*100 / len(sent_df),
        'adjusted_packet_loss_rate(%)': len(real_lost_df)*100 / len(sent_df)
    }]

    # Convert the dictionary to a dataframe
    statistics_df = pd.DataFrame.from_dict(statistics_data)
    return statistics_df

file_pairs = []
for date in dates:
    for exp in exp_names:
        for device in device_names:
            ##### ---------- READ TIME SYNC FILE ---------- #####
            sync_file_name = f"{database}/{date}/time_sync_{device}.json"
            with open(sync_file_name, 'r') as file:
                data = json.load(file)
            # Extract values from the dictionary
            time_diff_list = list(data.values())
            ##### ---------- READ TIME SYNC FILE ---------- #####

            ##### ---------- GET QLOG FILE ---------- #####
            find_pairs = FindFilePairs(database, date, exp, device)
            for pair in find_pairs:
               print(pair)
            print("TOTAL UL/DL PAIRS CNT:", len(find_pairs))
            ##### ---------- GET QLOG FILE ---------- #####
            for pair in find_pairs:
                print("NOW:", pair)
                ##### ---------- TRANSFORM QLOG ---------- #####
                time = pair[2]
                port = pair[3]
                exp_round = int(pair[5])
                sent_qlog_file = pair[0]
                received_qlog_file = pair[1]
                # convert to JSON entry
                sent_json_entry = QlogToJsonEntry(sent_qlog_file)
                received_json_entry = QlogToJsonEntry(received_qlog_file)
                # convert to .csv
                sent_csv_file = sent_qlog_file.replace(".qlog", ".csv")
                received_csv_file = received_qlog_file.replace(".qlog", ".csv")
                JsonToCsv(sent_json_entry, sent_csv_file)
                JsonToCsv(received_json_entry, received_csv_file)
                
                print("CONVERTING TO CSV COMPLETED.")
                ##### ---------- TRANSFORM QLOG ---------- #####

                sent_df = pd.read_csv(sent_csv_file)
                received_df = pd.read_csv(received_csv_file)

                ##### ---------- SYNC TIME TO SERVER TIME ---------- #####
                mean_time_diff = time_diff_list[exp_round-1] * 1000
                # No matter downlink or uplink, the file time that need to change is client side.
                if int(port)%2 == 0:    # UL
                    clientStartTime = GetStartTime(sent_json_entry)
                    print(clientStartTime)
                    serverStartTime = GetStartTime(received_json_entry)
                    print(serverStartTime)

                    senderRefTime = clientStartTime + mean_time_diff
                    rcverRefTime = serverStartTime

                else:   # DL
                    clientStartTime = GetStartTime(received_json_entry)
                    print(clientStartTime)
                    serverStartTime = GetStartTime(sent_json_entry)
                    print(serverStartTime)
                    startTimeDiff = (clientStartTime - serverStartTime) + mean_time_diff

                    senderRefTime = serverStartTime
                    rcverRefTime = clientStartTime + mean_time_diff
                
                # Add 8 hours to both epoch times and timestamps to match UMT+8
                # Also sync time with server
                sent_df = ProcessTime(sent_df, senderRefTime)
                epoch_times_gmt8 = sent_df["epoch_time"] + 8 * 3600 * 1000
                sent_df["epoch_time"] = epoch_times_gmt8
                timestamps_gmt8 = pd.to_datetime(epoch_times_gmt8, unit='ms').dt.strftime('%Y-%m-%d %H:%M:%S.%f')
                sent_df["timestamp"] = timestamps_gmt8
                
                received_df = ProcessTime(received_df, rcverRefTime)
                epoch_times_gmt8 = received_df["epoch_time"] + 8 * 3600 * 1000
                received_df["epoch_time"] = epoch_times_gmt8
                timestamps_gmt8 = pd.to_datetime(epoch_times_gmt8, unit='ms').dt.strftime('%Y-%m-%d %H:%M:%S.%f')
                received_df["timestamp"] = timestamps_gmt8


                ## Add RealTimeStamp to CSV
                new_csv_order = ['time', 'epoch_time', 'timestamp', 'name', 'data']
                sent_df = sent_df[new_csv_order]
                received_df = received_df[new_csv_order]

                sent_df.to_csv(sent_csv_file, index=False)
                received_df.to_csv(received_csv_file, index=False)
                ##### ---------- SYNC TIME TO SERVER TIME ---------- #####

                ##### ---------- PARSE THE DATA TYPE ---------- #####
                # sender side data
                metrics_all_rows = sent_df[(sent_df['name'] == 'recovery:metrics_updated') & (sent_df['data'].str.contains("'bytes_in_flight':"))]
                metrics_sent_rows = sent_df[(sent_df['name'] == 'recovery:metrics_updated') & (sent_df['data'].str.contains("{'bytes_in_flight':"))]
                metrics_ack_rows = sent_df[(sent_df['name'] == 'recovery:metrics_updated') & (sent_df['data'].str.contains("'latest_rtt':"))]
                total_sent_rows = sent_df[(sent_df['name'] == 'transport:packet_sent')]
                pk_sent_rows = sent_df[(sent_df['name'] == 'transport:packet_sent') & (sent_df['data'].str.contains("'frame_type': 'stream'"))]
                rcv_ack_rows = sent_df[(sent_df['name'] == 'transport:packet_received') & (sent_df['data'].str.contains("'frame_type': 'ack'")) & (sent_df['data'].str.contains("'packet_type': '1RTT'"))]
                lost_rows = sent_df[sent_df['name'] == 'recovery:packet_lost']

                # Get the count of rows
                pk_sent_cnt = len(pk_sent_rows)
                rcv_ack_cnt = len(rcv_ack_rows)

                # receiver side data
                pk_rcv_rows = received_df[(received_df['name'] == "transport:packet_received") & (received_df['data'].str.contains("'frame_type': 'stream'"))]
                pk_rcv_rows = pk_rcv_rows.reset_index(drop=True)
                ##### ---------- PARSE THE DATA TYPE ---------- #####

                ##### ---------- CONCAT `transport:packet_sent` & `recovery:metrics_updated` ---------- ######
                metrics_sent_rows = metrics_sent_rows.reset_index(drop=True)
                pk_sent_rows = pk_sent_rows.reset_index(drop=True)

                for i in range(pk_sent_cnt):
                    if(i >= len(metrics_sent_rows)):
                        data = metrics_sent_rows.iloc[i-1]['data']
                        new_row_data = {'time': [pk_sent_rows.iloc[i]['time']], 'name':['recovery:metrics_updated'], 'data': [data]}
                        new_row = pd.DataFrame(new_row_data)
                        metrics_sent_rows = pd.concat([metrics_sent_rows, new_row], ignore_index=True)
                        continue
                    time_diff = metrics_sent_rows.iloc[i]['time'] - pk_sent_rows.iloc[i]['time']
                    # time_diff >= 1: not the matching metrics_update
                    while time_diff >= 1:
                        data = metrics_sent_rows.iloc[i-1]['data']
                        new_row_data = {'time': [pk_sent_rows.iloc[i]['time']], 'name':['recovery:metrics_updated'], 'data': [data]}
                        new_row = pd.DataFrame(new_row_data)
                        metrics_sent_rows = insert(metrics_sent_rows, i, new_row)
                        time_diff = metrics_sent_rows.iloc[i]['time'] - pk_sent_rows.iloc[i]['time']
                    # time_diff < 0: missing metrics_update
                    while time_diff < 0:
                        # print(i, time_diff_list)
                        metrics_sent_rows.drop(index=metrics_sent_rows.index[i], inplace=True)
                        time_diff = metrics_sent_rows.iloc[i]['time'] - pk_sent_rows.iloc[i]['time']
                
                metrics_sent_rows = metrics_sent_rows.reset_index(drop=True)
                pk_sent_rows = pk_sent_rows.reset_index(drop=True)
                print("SENT ROWS CNT:", len(metrics_sent_rows), len(pk_sent_rows))

                # extract bytes_in_flight & packets_in_flight
                metrics_sent_rows['bytes_in_flight'] = None
                metrics_sent_rows['packets_in_flight'] = None
                # Use ast.literal_eval to safely evaluate the string and extract 'bytes_in_flight' and 'packets_in_flight'
                metrics_sent_rows[['bytes_in_flight', 'packets_in_flight']] = metrics_sent_rows['data'].apply(
                    lambda x: pd.Series(ast.literal_eval(x)) if isinstance(x, str) else pd.Series([None, None]))
                # Add bytes_in_flight & packets_in_flight to pk_sent_rows
                pk_sent_rows['bytes_in_flight'] = metrics_sent_rows['bytes_in_flight']
                pk_sent_rows['packets_in_flight'] = metrics_sent_rows['packets_in_flight']
                ##### ---------- CONCAT `transport:packet_sent` & `recovery:metrics_updated` ---------- ######
                        
                ##### --------- CONCAT `transport:packet_received` & `recovery:metrics_updated` ---------- #####
                metrics_ack_rows = metrics_ack_rows.reset_index(drop=True)
                rcv_ack_rows = rcv_ack_rows.reset_index(drop=True)
                initial_ack_metrics = metrics_ack_rows.iloc[[0]]
                metrics_ack_rows.drop(index=metrics_ack_rows.index[0], inplace=True)
                metrics_ack_rows = metrics_ack_rows.reset_index(drop=True)

                for i in range(rcv_ack_cnt):
                    if(i >= len(metrics_ack_rows)):
                        data = metrics_ack_rows.iloc[i-1]['data']
                        new_row_data = {'time': [rcv_ack_rows.iloc[i]['time']], 'name':['recovery:metrics_updated'], 'data': [data]}
                        new_row = pd.DataFrame(new_row_data)
                        metrics_ack_rows = pd.concat([metrics_ack_rows, new_row], ignore_index=True)
                        continue
                    time_diff = metrics_ack_rows.iloc[i]['time'] - rcv_ack_rows.iloc[i]['time']
                    # time_diff >= 1: missing metrics_update 
                    while time_diff > 0:
                        if i == 0:
                            data = initial_ack_metrics.iloc[0]['data']
                        else:
                            data = metrics_ack_rows.iloc[i-1]['data']
                        new_row_data = {'time': [rcv_ack_rows.iloc[i]['time']], 'name':['recovery:metrics_updated'], 'data': [data]}
                        new_row = pd.DataFrame(new_row_data)
                        metrics_ack_rows = insert(metrics_ack_rows, i, new_row)
                        time_diff = metrics_ack_rows.iloc[i]['time'] - rcv_ack_rows.iloc[i]['time']
                    # time_diff < 0: not the matching metrics_update
                    while time_diff <= -1:
                        metrics_ack_rows.drop(index=metrics_ack_rows.index[i], inplace=True)
                        time_diff = metrics_ack_rows.iloc[i]['time'] - rcv_ack_rows.iloc[i]['time']
                
                metrics_ack_rows = metrics_ack_rows.reset_index(drop=True)
                rcv_ack_rows = rcv_ack_rows.reset_index(drop=True)
                print("RECEIVED ROWS CNT:", len(metrics_ack_rows), len(rcv_ack_rows))

                ack_json_list = []
                # Add the initial_ack_metrics for temporary
                print("initial_ack_metrics:", initial_ack_metrics)
                metrics_ack_rows = pd.concat([initial_ack_metrics, metrics_ack_rows], axis=0).reset_index(drop=True)
                for i in range(len(metrics_ack_rows)):
                    s = metrics_ack_rows.iloc[i]['data'].replace("\'", "\"")
                    json_object = json.loads(s)
                    ack_json_list.append(json_object)

                # Create a DataFrame for ACK info
                metrics_ack_df = pd.DataFrame(ack_json_list)
                # Fill missing values in each row with the previous row's values
                metrics_ack_df = metrics_ack_df.ffill(axis=0)

                # Drop initial_ack_metrics
                metrics_ack_rows.drop(index=metrics_ack_rows.index[0], inplace=True)
                metrics_ack_rows = metrics_ack_rows.reset_index(drop=True)
                metrics_ack_df.drop(index=metrics_ack_df.index[0], inplace=True)
                metrics_ack_df = metrics_ack_df.reset_index(drop=True)
                metrics_ack_rows = pd.concat([metrics_ack_rows, metrics_ack_df], axis=1).reset_index(drop=True)
                # Since we have parse out all the information in data, we can drop the data column
                metrics_ack_rows = metrics_ack_rows.drop(columns=['data'])

                rcv_ack_rows = pd.concat([rcv_ack_rows, metrics_ack_df], axis=1)
                rcv_ack_rows = rcv_ack_rows.reset_index(drop=True)
                ##### ---------- CONCAT `transport:packet_received` & `recovery:metrics_updated` ---------- #####

                ##### ---------- MAPPING ACK: "packet_number", "offset", "length" ---------- #####
                acked_ranges_series = rcv_ack_rows['data']
                acked_ranges_list = []
                for i in range(len(acked_ranges_series)):
                    s = acked_ranges_series.iloc[i]
                    data_dict = json.loads(s.replace("\'", "\""))
                    # Extract 'acked_ranges' from all frames
                    acked_ranges = [range_entry for frame in data_dict['frames'] if 'acked_ranges' in frame for range_entry in frame['acked_ranges']]
                    acked_ranges_list.append(acked_ranges)

                acked_ranges_df = pd.DataFrame({"acked_ranges": acked_ranges_list})
                
                rcv_ack_rows = pd.concat([rcv_ack_rows, acked_ranges_df], axis=1)
                rcv_ack_rows = rcv_ack_rows.reset_index(drop=True)

                # Parse out the packet_number & offset & length
                pk_sent_series =  pk_sent_rows['data']
                pk_num_list = []
                offset_list = []
                length_list = []
                for i in range(len(pk_sent_series)):
                    s = pk_sent_series.iloc[i]
                    data_dict = json.loads(s.replace("\'", "\""))
                    packet_number = data_dict['header']['packet_number']
                    # Initialize offset to None in case 'frame_type': 'stream' is not found
                    offset = None
                    # Iterate through frames to find 'offset' for 'frame_type': 'stream'
                    for frame in data_dict.get('frames', []):
                        if frame.get('frame_type') == 'stream':
                            offset = frame.get('offset')
                            length = frame.get('length')
                            break  # Stop iterating once 'offset' is found
                    
                    pk_num_list.append(packet_number)
                    offset_list.append(offset)
                    length_list.append(length)

                pk_num_df = pd.DataFrame({"packet_number": pk_num_list, "offset": offset_list, "length": length_list})

                pk_sent_rows = pd.concat([pk_sent_rows, pk_num_df], axis=1)
                pk_sent_rows = pk_sent_rows.reset_index(drop=True)
                ##### ---------- MAPPING ACK: "packet_number", "offset", "length" ---------- #####

                ##### ---------- MAPPING ACK: "smoothed_rtt", "latest_rtt", "congestion_window", "rtt_variance" ---------- #####
                print("Start parsing RTT info")
                t = TicToc()
                t.tic()
                pk_sent_rows['smoothed_rtt'] = np.nan
                pk_sent_rows['latest_rtt'] = np.nan
                pk_sent_rows['rtt_variance'] = np.nan
                pk_sent_rows['congestion_window'] = np.nan
                
                # Create a dictionary to store the mapping between packet numbers and indices
                packet_number_index_map = {packet_number: index for index, packet_number in enumerate(pk_sent_rows['packet_number'])}
                # Apply the custom update function to each row in rcv_ack_rows
                for idx, row in rcv_ack_rows.iterrows():
                    update_pk_sent_rows(row)
                # rcv_ack_rows.apply(update_pk_sent_rows, axis=1)
                t.toc("parsing RTT info took")
                ##### ---------- MAPPING ACK: "smoothed_rtt", "latest_rtt", "congestion_window", "rtt_variance" ---------- #####

                ##### ---------- IDENTIFY LOST PACKETS ---------- #####
                # Use ast.literal_eval to safely evaluate the string and extract 'packet_number'
                lost_rows['packet_number'] = lost_rows['data'].apply(lambda x: ast.literal_eval(x)['header']['packet_number'] if isinstance(x, str) else None)
                lost_rows['trigger'] = lost_rows['data'].apply(lambda x: ast.literal_eval(x)['trigger'] if isinstance(x, str) else None)
                # Export packet lost file
                lost_pk_csv_file_path = f"{database}/{date}/{exp}/{device}/#{pair[5]}/middle/lost_pk_{time}_{port}.csv"
                lost_rows.to_csv(lost_pk_csv_file_path, index=False)

                # Set to True if the packet is lost
                pk_sent_rows['packet_lost'] = False

                # Iterate through rows and set 'packet_lost' to True where 'packet_number' values match
                for _, lost_row in lost_rows.iterrows():
                    packet_number = lost_row['packet_number']
                    
                    # Check if 'packet_number' exists in pk_sent_rows
                    if packet_number in pk_sent_rows['packet_number'].values:
                        pk_sent_rows.loc[pk_sent_rows['packet_number'] == packet_number, 'packet_lost'] = True
                ##### ---------- IDENTIFY LOST PACKETS ---------- #####
                        
                ##### ---------- PROCESSED SENT FILE ---------- #####
                cols = ['time', 'epoch_time', 'timestamp', 'name', 'packet_number', 'offset', 'length', 'bytes_in_flight', 'packets_in_flight', 'smoothed_rtt', 'latest_rtt', 'rtt_variance', 'congestion_window', 'packet_lost', 'data']
                processed_sent_df = pk_sent_rows[cols]
                csv_file_path = f"{database}/{date}/{exp}/{device}/#{pair[5]}/data/processed_sent_{time}_{port}.csv"
                processed_sent_df.to_csv(csv_file_path, sep='@', index=False)
                ##### ---------- PROCESSED SENT FILE ---------- #####

                ##### ---------- RECEIVER SIDE DATA ---------- #####
                pk_rcv_df = pk_rcv_rows.reset_index(drop=True)

                pk_rcv_series =  pk_rcv_df['data']
                pk_rcv_num_list = []
                offset_rcv_list = []
                length_rcv_list = []
                for i in range(len(pk_rcv_series)):
                    s = pk_rcv_series.iloc[i]
                    data_dict = json.loads(s.replace("\'", "\""))
                    packet_number = data_dict['header']['packet_number']
                    # Initialize offset to None in case 'frame_type': 'stream' is not found
                    offset = None
                    # Iterate through frames to find 'offset' for 'frame_type': 'stream'
                    for frame in data_dict.get('frames', []):
                        if frame.get('frame_type') == 'stream':
                            offset = frame.get('offset')
                            length = frame.get('length')
                            break  # Stop iterating once 'offset' is found
                    
                    pk_rcv_num_list.append(packet_number)
                    offset_rcv_list.append(offset)
                    length_rcv_list.append(length)

                pk_rcv_df['packet_number'] = pk_rcv_num_list
                pk_rcv_df['offset'] = offset_rcv_list
                pk_rcv_df['length'] = length_rcv_list
                ##### ---------- RECEIVER SIDE DATA ---------- #####

                ##### ---------- PROCESSED RECEIVED FILE ---------- #####
                cols = ['time', 'epoch_time', 'timestamp', 'name', 'packet_number', 'offset', 'length', 'data']
                processed_rcv_df = pk_rcv_df[cols]
                csv_file_path = f"{database}/{date}/{exp}/{device}/#{pair[5]}/data/processed_rcv_{time}_{port}.csv"
                processed_rcv_df.to_csv(csv_file_path, sep='@')
                ##### ---------- PROCESSED RECEIVED FILE ---------- #####


##### ---------- CALCULATE LOST PACKET STATISTICS ---------- #####
all_ul_sender_files = []
all_ul_rcv_files = []
all_ul_pkl_files = []
for date in dates:
    for exp in exp_names:
        for device in device_names:
            ul_sender_files = find_ul_sender_file(database, date, exp, device)
            ul_rcv_files = find_ul_rcv_file(database, date, exp, device)
            ul_pk_loss_files = find_ul_loss_file(database, date, exp, device)
            all_ul_sender_files.extend(ul_sender_files)
            all_ul_rcv_files.extend(ul_rcv_files)
            all_ul_pkl_files.extend(ul_pk_loss_files)

for i in range(len(all_ul_rcv_files)):
    ul_sender_df = pd.read_csv(all_ul_sender_files[i], sep=',')
    # ul_sent_df is raw file, ul_rcv_df is processed file
    ul_sent_df = ul_sender_df[(ul_sender_df['name'] == 'transport:packet_sent')]
    ul_data_df = ul_sent_df[ul_sent_df['data'].str.contains("'frame_type': 'stream'")]
    ul_rcv_df = pd.read_csv(all_ul_rcv_files[i], sep='@')
    ul_loss_df = pd.read_csv(all_ul_pkl_files[i])
    ul_exec_lat_df, ul_exec_reorder_df, ul_exec_time_df, ul_real_lost_df, ul_lost_reorder_df, ul_lost_time_df = get_loss_data(ul_loss_df, ul_rcv_df)
    ul_statistics = calculate_statistics(ul_lost_reorder_df, ul_lost_time_df, ul_real_lost_df, ul_exec_reorder_df, ul_exec_time_df, ul_exec_lat_df, ul_loss_df, ul_data_df, ul_sent_df)

    directory = os.path.dirname(all_ul_rcv_files[i])
    ul_loss_df['lost'] = False
    ul_loss_df['excl'] = False
    # Set 'lost' column to True for rows in ul_real_lost_df
    ul_loss_df.loc[ul_loss_df['packet_number'].isin(ul_real_lost_df['packet_number']), 'lost'] = True
    # Set 'excl' column to True for rows in ul_exec_lat_df
    ul_loss_df.loc[ul_loss_df['packet_number'].isin(ul_exec_lat_df['packet_number']), 'excl'] = True
    parts = directory.split("/")
    parts[-1] = "data"
    data_directory = "/".join(parts)
    ul_loss_df.to_csv(f"{data_directory}/ul_real_lost_pk.csv", index=False)
    # modify to the statistics directory
    parts = directory.split("/")
    parts[-1] = "statistics"
    statistics_directory = "/".join(parts)
    ul_statistics.to_csv(f"{statistics_directory}/ul_statistics.csv", index=False)

all_dl_sender_files = []
all_dl_rcv_files = []
all_dl_pkl_files = []
for date in dates:
    for exp in exp_names:
        for device in device_names:
            dl_sender_files = find_dl_sender_file(database, date, exp, device)
            dl_rcv_files = find_dl_rcv_file(database, date, exp, device)
            dl_pk_loss_files = find_dl_loss_file(database, date, exp, device)
            all_dl_sender_files.extend(dl_sender_files)
            all_dl_rcv_files.extend(dl_rcv_files)
            all_dl_pkl_files.extend(dl_pk_loss_files)

for i in range(len(all_dl_rcv_files)):
    dl_sender_df = pd.read_csv(all_dl_sender_files[i], sep=',')
    # dl_sent_df is raw file, dl_rcv_df is processed file
    dl_sent_df = dl_sender_df[(dl_sender_df['name'] == 'transport:packet_sent')]
    dl_data_df = dl_sent_df[dl_sent_df['data'].str.contains("'frame_type': 'stream'")]
    dl_rcv_df = pd.read_csv(all_dl_rcv_files[i], sep='@')
    dl_loss_df = pd.read_csv(all_dl_pkl_files[i])
    dl_exec_lat_df, dl_exec_reorder_df, dl_exec_time_df, dl_real_lost_df, dl_lost_reorder_df, dl_lost_time_df = get_loss_data(dl_loss_df, dl_rcv_df)
    dl_statistics = calculate_statistics(dl_lost_reorder_df, dl_lost_time_df, dl_real_lost_df, dl_exec_reorder_df, dl_exec_time_df, dl_exec_lat_df, dl_loss_df, dl_data_df, dl_sent_df)
    
    directory = os.path.dirname(all_dl_rcv_files[i])
    dl_loss_df['lost'] = False
    dl_loss_df['excl'] = False
    # Set 'lost' column to True for rows in dl_real_lost_df
    dl_loss_df.loc[dl_loss_df['packet_number'].isin(dl_real_lost_df['packet_number']), 'lost'] = True
    # Set 'excl' column to True for rows in ul_exec_lat_df
    dl_loss_df.loc[dl_loss_df['packet_number'].isin(dl_exec_lat_df['packet_number']), 'excl'] = True
    dl_loss_df.to_csv(f"{directory}/dl_real_lost_pk.csv", index=False)
    # modify to the statistics directory
    parts = directory.split("/")
    parts[-1] = "statistics"
    statistics_directory = "/".join(parts)
    dl_statistics.to_csv(f"{statistics_directory}/dl_statistics.csv", index=False)
##### ---------- CALCULATE LOST PACKET STATISTICS ---------- #####