# undo.go

## Introduce

undo.go is a demostration of how undo log can be implemented. It DOES NOT support concurrent calling.

### Data structure

It generate a binary log file with a header and followed with items. Items are appened to the file in the order that they are generated.
Header consists of:

- A magic number, to identify our log file.
- Version information
- Offset of thet first item recorded
- Offset of the last item recorded
- Size of the file

Each items consist of:

- Type of item(write|commit)
- Offset of the previous item
- Offset of the next item
- Information of transaction associated
- Information of user from whom cash is transact
- Information of user to whom cash is transact

So items can be retrieved from beginning or from the end. Any call to Write() or Pop() will be written to file synchronously.

### Transaction

Normal transactions will have a write type of item followed by a commit type of item. Caller should call Write() with a write type BEFORE update the value of both accounts, in case of system failure, and call Write() again with a commit type of item right after the transaction is done.

### Undo transactions

To undo the last transaction, call Pop(). To undo a transaction from a certain ID, call Read() to retrieve the last tranaction and then Pop(), repeatly, until you get the very item. The write and commit type of items will be handled in pairs within a single call.

### Recovery

If file is not closed properly, header may not be updated, then recovery will be performed in the next openning: offset of the last item and size of the file will be updated and written to file header again. Caller should try to undo the last transaction, i.e. the one with out a commit item. Errors will be returned if recovery fail.

### Limitation

File size over 2GB is not supported. Don't do that!
Caller should ensure that the transaction id is valid before performing undo.

## Usage

undoLog := NewUndoLog("./undo.bin")
undoLog.Write(&UndoItem(Cmd:write,
        TranscationID: transcationID,
        FromID:        fromID,
        FromCash:      fromCash,
        ToID:          toID,
        ToCash:        toCash,
        Cash:          cash,
        ))
//Update account info here
...
undoLog.Write(&UndoItem{Cmd: commit, TranscationID: t.TranscationID})

//this maybe in a loop
item := undoLog.Read()
//test item.transactionid
...
undoLog.Pop()
//loop end

undoLog.Close()
