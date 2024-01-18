'''
from: https://github.com/Jackbedford0428/wmnl-handoff-research/blob/main/data-preprocessing-beta/mi2log/mi_offline_analysis.py
'''
import os
import sys
import argparse
import traceback
from pprint import pprint
from pytictoc import TicToc

### Import MobileInsight modules
from mobile_insight.monitor import OfflineReplayer
from mobile_insight.analyzer import MsgLogger, NrRrcAnalyzer, LteRrcAnalyzer, WcdmaRrcAnalyzer, LteNasAnalyzer, UmtsNasAnalyzer, LteMacAnalyzer, LtePhyAnalyzer, LteMeasurementAnalyzer

# ******************************* User Settings *******************************
database = "/home/wmnlab/Desktop/"
dates = [
         "2024-01-17",
]
devices = sorted([
    "sm00",
    "sm01",
])
exps = {  # experiment_name: (number_of_experiment_rounds, list_of_experiment_round)
            # If the list is None, it will not list as directories.
            # If the list is empty, it will list all directories in the current directory by default.
            # If the number of experiment times != the length of existing directories of list, it would trigger warning and skip the directory.
    "pm": (2, ["#{:02d}".format(i+1) for i in range(2)]),
}
# *****************************************************************************

# **************************** Auxiliary Functions ****************************
def mi_decode(fin, fout):
    try:
        ### Initialize a monitor
        src = OfflineReplayer()
        # src.set_input_path("./logs/")
        src.set_input_path(fin)
        src.enable_log_all()

        # src.enable_log("LTE_PHY_Serv_Cell_Measurement")
        # src.enable_log("5G_NR_RRC_OTA_Packet")
        # src.enable_log("LTE_RRC_OTA_Packet")
        # src.enable_log("LTE_NB1_ML1_GM_DCI_Info")

        logger = MsgLogger()
        logger.set_decode_format(MsgLogger.XML)
        logger.set_dump_type(MsgLogger.FILE_ONLY)
        logger.save_decoded_msg_as(fout)
        logger.set_source(src)

        ### Analyzers
        nr_rrc_analyzer = NrRrcAnalyzer()
        nr_rrc_analyzer.set_source(src)  # bind with the monitor

        lte_rrc_analyzer = LteRrcAnalyzer()
        lte_rrc_analyzer.set_source(src)  # bind with the monitor

        wcdma_rrc_analyzer = WcdmaRrcAnalyzer()
        wcdma_rrc_analyzer.set_source(src)  # bind with the monitor

        # lte_nas_analyzer = LteNasAnalyzer()
        # lte_nas_analyzer.set_source(src)

        # umts_nas_analyzer = UmtsNasAnalyzer()
        # umts_nas_analyzer.set_source(src)

        lte_mac_analyzer = LteMacAnalyzer()
        lte_mac_analyzer.set_source(src)

        lte_phy_analyzer = LtePhyAnalyzer()
        lte_phy_analyzer.set_source(src)

        lte_meas_analyzer = LteMeasurementAnalyzer()
        lte_meas_analyzer.set_source(src)

        # print lte_meas_analyzer.get_rsrp_list() 
        # print lte_meas_analyzer.get_rsrq_list()

        ### Start the monitoring
        src.run()
    except:
        ### Record error message without halting the program
        return (fin, fout, traceback.format_exc())
    return (fin, fout, None)
# *****************************************************************************

# **************************** Utils Functions ****************************
def makedir(dirpath, mode=0):  # mode=1: show message, mode=0: hide message
    if os.path.isdir(dirpath):
        if mode:
            print("mkdir: cannot create directory '{}': directory has already existed.".format(dirpath))
        return
    ### recursively make directory
    _temp = []
    while not os.path.isdir(dirpath):
        _temp.append(dirpath)
        dirpath = os.path.dirname(dirpath)
    while _temp:
        dirpath = _temp.pop()
        print("mkdir", dirpath)
        os.mkdir(dirpath)

def error_handling(err_handle):
    """
    Print the error messages during the process.
    
    Args:
        err_handle (str-tuple): (input_filename, output_filename, error_messages : traceback.format_exc())
    Returns:
        (bool): check if the error_messages occurs, i.e., whether it is None.
    """
    if err_handle[2]:
        print()
        print("**************************************************")
        print("File decoding from '{}' into '{}' was interrupted.".format(err_handle[0], err_handle[1]))
        print()
        print(err_handle[2])
        return True
    return False
# *****************************************************************************

if __name__ == "__main__":
    def fgetter():
        files_collection = []
        tags = "diag_log"
        for filename in filenames:
            if filename.startswith(tags) and filename.endswith(".mi2log"):
                files_collection.append(filename)
        return files_collection
    
    def main():
        files_collection = fgetter()
        if len(files_collection) == 0:
            print("No candidate file.")
        for filename in files_collection:
            fin = os.path.join(source_dir, filename)
            fout = os.path.join(target_dir, "{}.txt".format(filename[:-7]))
            print(">>>>> decode from '{}' into '{}'...".format(fin, fout))
            err_handle = mi_decode(fin, fout)
            err_handles.append(err_handle)
        print()
    
    # ******************************* Check Files *********************************
    for date in dates:
        for expr, (times, traces) in exps.items():
            print(os.path.join(database, date, expr))
            for dev in devices:
                if not os.path.isdir(os.path.join(database, date, expr, dev)):
                    print("|___ {} does not exist.".format(os.path.join(database, date, expr, dev)))
                    continue
                
                print("|___", os.path.join(database, date, expr, dev))
                if traces == None:
                    # print(os.path.join(database, date, expr, dev))
                    continue
                elif len(traces) == 0:
                    traces = sorted(os.listdir(os.path.join(database, date, expr, dev)))
                
                print("|    ", times)
                traces = [trace for trace in traces if os.path.isdir(os.path.join(database, date, expr, dev, trace))]
                if len(traces) != times:
                    print("***************************************************************************************")
                    print("Warning: the number of traces does not match the specified number of experiment times.")
                    print("***************************************************************************************")
                for trace in traces:
                    print("|    |___", os.path.join(database, date, expr, dev, trace))
            print()
    # *****************************************************************************

    # ******************************** Processing *********************************
    t = TicToc()  # create instance of class
    t.tic()       # Start timer
    err_handles = []
    for date in dates:
        for expr, (times, traces) in exps.items():
            for dev in devices:
                if not os.path.isdir(os.path.join(database, date, expr, dev)):
                    print("{} does not exist.\n".format(os.path.join(database, date, expr, dev)))
                    continue

                if traces == None:
                    print("------------------------------------------")
                    print(date, expr, dev)
                    print("------------------------------------------")
                    source_dir = os.path.join(database, date, expr, dev)
                    target_dir = os.path.join(database, date, expr, dev)
                    makedir(target_dir)
                    filenames = os.listdir(source_dir)
                    main()
                    continue
                elif len(traces) == 0:
                    traces = sorted(os.listdir(os.path.join(database, date, expr, dev)))
                
                traces = [trace for trace in traces if os.path.isdir(os.path.join(database, date, expr, dev, trace))]
                for trace in traces:
                    print("------------------------------------------")
                    print(date, expr, dev, trace)
                    print("------------------------------------------")
                    source_dir = os.path.join(database, date, expr, dev, trace, "raw")
                    target_dir = os.path.join(database, date, expr, dev, trace, "middle")
                    makedir(target_dir)
                    filenames = os.listdir(source_dir)
                    main()
    ### Check errors
    flag = False
    for err_handle in err_handles:
        flag = error_handling(err_handle)
    if not flag and err_handles:
        print("**************************************************")
        print("No error occurs!!")
        print("**************************************************")
    t.toc()  # Time elapsed since t.tic()
    # *****************************************************************************
